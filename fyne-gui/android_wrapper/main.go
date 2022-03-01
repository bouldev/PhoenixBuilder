package main

import (
	"android_pack/binres"
	"archive/zip"
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"encoding/xml"
	"fmt"
	"hash"
	"html/template"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"
	"strconv"
)

func main() {
	libFiles := []string{"lib/arm64-v8a/libphoenixbuilder-android-arm64.so"}
	libFilesSrc := []string{"build/phoenixbuilder-android-arm64-lib.so"}
	targetSDK := 30
	iconPath := "android_wrapper/icon.png"
	bundleID := "pro.fastbuilder.app"
	libname := "phoenixbuilder-android-arm64"// a~z A~Z 0~9 _
	appName := "phoenixbuilder-android-app"
	outputFile := "build/" + appName + ".apk"
	version := os.Args[1] //"0.0.4"
	build, _ := strconv.Atoi(os.Args[2]) //200

	// AndroidManifest.xml
	buf := new(bytes.Buffer)
	buf.WriteString(`<?xml version="1.0" encoding="utf-8"?>`)
	getManifestTmpl().Execute(buf, manifestTmplData{
		JavaPkgPath: bundleID,
		Name:        strings.Title(appName),
		Debug:       false,
		LibName:     libname,
		Version:     version,
		Build:       build,
	})
	manifestData := buf.Bytes()

	out, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}

	block, _ := pem.Decode([]byte(debugCert))
	privKey, _ := x509.ParsePKCS1PrivateKey(block.Bytes)
	apkWriter := NewWriter(out, privKey)
	dexWriter, _ := apkWriter.Create("classes.dex")
	dexData, _ := base64.StdEncoding.DecodeString(DexStr)
	dexWriter.Write(dexData)
	for i, libFile := range libFiles {
		w, _ := apkWriter.Create(libFile)
		src := libFilesSrc[i]
		f, err := os.Open(filepath.Clean(src))
		if err != nil {
			panic(err)
		}
		defer f.Close()
		if _, err := io.Copy(w, f); err != nil {
			panic(err)
		}
	}
	bxml, err := binres.UnmarshalXML(bytes.NewReader(manifestData), true, targetSDK)
	if err != nil {
		panic(err)
	}
	pkgname, err := bxml.RawValueByName("manifest", xml.Name{Local: "package"})
	if err != nil {
		panic(err)
	}
	tbl, name := binres.NewMipmapTable(pkgname)
	iconWriter, _ := apkWriter.Create(name)
	f, err := os.Open(filepath.Clean(iconPath))
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err := io.Copy(iconWriter, f); err != nil {
		panic(err)
	}
	w, _ := apkWriter.Create("resources.arsc")
	bin, err := tbl.MarshalBinary()
	if err != nil {
		panic(err)
	}
	if _, err := w.Write(bin); err != nil {
		panic(err)
	}

	w, err = apkWriter.Create("AndroidManifest.xml")
	if err != nil {
		panic(err)
	}
	bin, err = bxml.MarshalBinary()
	if err != nil {
		panic(err)
	}
	if _, err := w.Write(bin); err != nil {
		panic(err)
	}
	if err := apkWriter.Close(); err != nil {
		panic(err)
	}
	return
}

// NewWriter returns a new Writer writing an APK file to w.
// The APK will be signed with key.
func NewWriter(w io.Writer, priv *rsa.PrivateKey) *Writer {
	apkw := &Writer{priv: priv}
	apkw.w = zip.NewWriter(&countWriter{apkw: apkw, w: w})
	return apkw
}

type countWriter struct {
	apkw *Writer
	w    io.Writer
}

func (c *countWriter) Write(p []byte) (n int, err error) {
	n, err = c.w.Write(p)
	c.apkw.offset += n
	return n, err
}

// Writer implements an APK file writer.
type Writer struct {
	offset   int
	w        *zip.Writer
	priv     *rsa.PrivateKey
	manifest []manifestEntry
	cur      *fileWriter
}

// Create adds a file to the APK archive using the provided name.
//
// The name must be a relative path. The file's contents must be written to
// the returned io.Writer before the next call to Create or Close.
func (w *Writer) Create(name string) (io.Writer, error) {
	if err := w.clearCur(); err != nil {
		return nil, fmt.Errorf("apk: Create(%s): %v", name, err)
	}
	res, err := w.create(name)
	if err != nil {
		return nil, fmt.Errorf("apk: Create(%s): %v", name, err)
	}
	return res, nil
}

func (w *Writer) create(name string) (io.Writer, error) {
	// Align start of file contents by using Extra as padding.
	if err := w.w.Flush(); err != nil { // for exact offset
		return nil, err
	}
	const fileHeaderLen = 30 // + filename + extra
	start := w.offset + fileHeaderLen + len(name)
	extra := start % 4

	zipfw, err := w.w.CreateHeader(&zip.FileHeader{
		Name:  name,
		Extra: make([]byte, extra),
	})
	if err != nil {
		return nil, err
	}
	w.cur = &fileWriter{
		name: name,
		w:    zipfw,
		sha1: sha1.New(),
	}
	return w.cur, nil
}

// Close finishes writing the APK. This includes writing the manifest and
// signing the archive, and writing the ZIP central directory.
//
// It does not close the underlying writer.
func (w *Writer) Close() error {
	if err := w.clearCur(); err != nil {
		return fmt.Errorf("apk: %v", err)
	}

	hasDex := false
	for _, entry := range w.manifest {
		if entry.name == "classes.dex" {
			hasDex = true
			break
		}
	}

	manifest := new(bytes.Buffer)
	if hasDex {
		fmt.Fprint(manifest, manifestDexHeader)
	} else {
		fmt.Fprint(manifest, manifestHeader)
	}
	certBody := new(bytes.Buffer)

	for _, entry := range w.manifest {
		n := entry.name
		h := base64.StdEncoding.EncodeToString(entry.sha1.Sum(nil))
		fmt.Fprintf(manifest, "Name: %s\nSHA1-Digest: %s\n\n", n, h)
		cHash := sha1.New()
		fmt.Fprintf(cHash, "Name: %s\r\nSHA1-Digest: %s\r\n\r\n", n, h)
		ch := base64.StdEncoding.EncodeToString(cHash.Sum(nil))
		fmt.Fprintf(certBody, "Name: %s\nSHA1-Digest: %s\n\n", n, ch)
	}

	mHash := sha1.New()
	_, err := mHash.Write(manifest.Bytes())
	if err != nil {
		return err
	}
	cert := new(bytes.Buffer)
	fmt.Fprint(cert, certHeader)
	fmt.Fprintf(cert, "SHA1-Digest-Manifest: %s\n\n", base64.StdEncoding.EncodeToString(mHash.Sum(nil)))
	cert.Write(certBody.Bytes())

	mw, err := w.Create("META-INF/MANIFEST.MF")
	if err != nil {
		return err
	}
	if _, err := mw.Write(manifest.Bytes()); err != nil {
		return err
	}

	cw, err := w.Create("META-INF/CERT.SF")
	if err != nil {
		return err
	}
	if _, err := cw.Write(cert.Bytes()); err != nil {
		return err
	}

	rsa, err := signPKCS7(rand.Reader, w.priv, cert.Bytes())
	if err != nil {
		return fmt.Errorf("apk: %v", err)
	}
	rw, err := w.Create("META-INF/CERT.RSA")
	if err != nil {
		return err
	}
	if _, err := rw.Write(rsa); err != nil {
		return err
	}

	return w.w.Close()
}

const manifestHeader = `Manifest-Version: 1.0
Created-By: 1.0 (Go)

`

const manifestDexHeader = `Manifest-Version: 1.0
Dex-Location: classes.dex
Created-By: 1.0 (Go)

`

const certHeader = `Signature-Version: 1.0
Created-By: 1.0 (Go)
`

func (w *Writer) clearCur() error {
	if w.cur == nil {
		return nil
	}
	w.manifest = append(w.manifest, manifestEntry{
		name: w.cur.name,
		sha1: w.cur.sha1,
	})
	w.cur.closed = true
	w.cur = nil
	return nil
}

type manifestEntry struct {
	name string
	sha1 hash.Hash
}

type fileWriter struct {
	name   string
	w      io.Writer
	sha1   hash.Hash
	closed bool
}

func (w *fileWriter) Write(p []byte) (n int, err error) {
	if w.closed {
		return 0, fmt.Errorf("apk: write to closed file %q", w.name)
	}
	_, err = w.sha1.Write(p)
	if err != nil {
		return 0, fmt.Errorf("apk: sha1 write %s", err)
	}
	n, err = w.w.Write(p)
	if err != nil {
		return 0, fmt.Errorf("apk: %v", err)
	}
	return n, err
}

type manifestTmplData struct {
	JavaPkgPath string
	Name        string
	Debug       bool
	LibName     string
	Version     string
	Build       int
}

func getManifestTmpl() *template.Template {
	t := template.Must(template.New("manifest").Parse(`
<manifest
	xmlns:android="http://schemas.android.com/apk/res/android"
	package="{{.JavaPkgPath}}"
	android:versionCode="{{.Build}}"
	android:versionName="{{.Version}}">

	<application android:label="{{.Name}}" android:debuggable="{{.Debug}}">
	<activity android:name="org.golang.app.GoNativeActivity"
		android:label="{{.Name}}"
		android:configChanges="orientation|keyboardHidden|uiMode"
		android:theme="@android:style/Theme">
		<meta-data android:name="android.app.lib_name" android:value="{{.LibName}}" />
		<intent-filter>
			<action android:name="android.intent.action.MAIN" />
			<category android:name="android.intent.category.LAUNCHER" />
		</intent-filter>
	</activity>
	</application>

	<uses-permission android:name="android.permission.WRITE_EXTERNAL_STORAGE" />
	<uses-permission android:name="android.permission.READ_EXTERNAL_STORAGE" />
	<uses-permission android:name="android.permission.INTERNET" />
</manifest>`))
	return t
}

// A random uninteresting private key.
// Must be consistent across builds so newer app versions can be installed.
const debugCert = `
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAy6ItnWZJ8DpX9R5FdWbS9Kr1U8Z7mKgqNByGU7No99JUnmyu
NQ6Uy6Nj0Gz3o3c0BXESECblOC13WdzjsH1Pi7/L9QV8jXOXX8cvkG5SJAyj6hcO
LOapjDiN89NXjXtyv206JWYvRtpexyVrmHJgRAw3fiFI+m4g4Qop1CxcIF/EgYh7
rYrqh4wbCM1OGaCleQWaOCXxZGm+J5YNKQcWpjZRrDrb35IZmlT0bK46CXUKvCqK
x7YXHgfhC8ZsXCtsScKJVHs7gEsNxz7A0XoibFw6DoxtjKzUCktnT0w3wxdY7OTj
9AR8mobFlM9W3yirX8TtwekWhDNTYEu8dwwykwIDAQABAoIBAA2hjpIhvcNR9H9Z
BmdEecydAQ0ZlT5zy1dvrWI++UDVmIp+Ve8BSd6T0mOqV61elmHi3sWsBN4M1Rdz
3N38lW2SajG9q0fAvBpSOBHgAKmfGv3Ziz5gNmtHgeEXfZ3f7J95zVGhlHqWtY95
JsmuplkHxFMyITN6WcMWrhQg4A3enKLhJLlaGLJf9PeBrvVxHR1/txrfENd2iJBH
FmxVGILL09fIIktJvoScbzVOneeWXj5vJGzWVhB17DHBbANGvVPdD5f+k/s5aooh
hWAy/yLKocr294C4J+gkO5h2zjjjSGcmVHfrhlXQoEPX+iW1TGoF8BMtl4Llc+jw
lKWKfpECgYEA9C428Z6CvAn+KJ2yhbAtuRo41kkOVoiQPtlPeRYs91Pq4+NBlfKO
2nWLkyavVrLx4YQeCeaEU2Xoieo9msfLZGTVxgRlztylOUR+zz2FzDBYGicuUD3s
EqC0Wv7tiX6dumpWyOcVVLmR9aKlOUzA9xemzIsWUwL3PpyONhKSq7kCgYEA1X2F
f2jKjoOVzglhtuX4/SP9GxS4gRf9rOQ1Q8DzZhyH2LZ6Dnb1uEQvGhiqJTU8CXxb
7odI0fgyNXq425Nlxc1Tu0G38TtJhwrx7HWHuFcbI/QpRtDYLWil8Zr7Q3BT9rdh
moo4m937hLMvqOG9pyIbyjOEPK2WBCtKW5yabqsCgYEAu9DkUBr1Qf+Jr+IEU9I8
iRkDSMeusJ6gHMd32pJVCfRRQvIlG1oTyTMKpafmzBAd/rFpjYHynFdRcutqcShm
aJUq3QG68U9EAvWNeIhA5tr0mUEz3WKTt4xGzYsyWES8u4tZr3QXMzD9dOuinJ1N
+4EEumXtSPKKDG3M8Qh+KnkCgYBUEVSTYmF5EynXc2xOCGsuy5AsrNEmzJqxDUBI
SN/P0uZPmTOhJIkIIZlmrlW5xye4GIde+1jajeC/nG7U0EsgRAV31J4pWQ5QJigz
0+g419wxIUFryGuIHhBSfpP472+w1G+T2mAGSLh1fdYDq7jx6oWE7xpghn5vb9id
EKLjdwKBgBtz9mzbzutIfAW0Y8F23T60nKvQ0gibE92rnUbjPnw8HjL3AZLU05N+
cSL5bhq0N5XHK77sscxW9vXjG0LJMXmFZPp9F6aV6ejkMIXyJ/Yz/EqeaJFwilTq
Mc6xR47qkdzu0dQ1aPm4XD7AWDtIvPo/GG2DKOucLBbQc2cOWtKS
-----END RSA PRIVATE KEY-----
`

var DexStr = `ZGV4CjAzNQCuiXhXcB/3x18IxZwivxbbtahJpJ5Vh42MKQAAcAAAAHhWNBIAAAAAAAAAAL` +
	`woAADMAAAAcAAAAC8AAACgAwAAQgAAAFwEAAAVAAAAdAcAAG4AAAAcCAAABgAAAIwLAABA` +
	`HQAATAwAADQYAAA2GAAAOxgAAD4YAABGGAAAWhgAAHEYAACBGAAAkRgAAJcYAACuGAAAsR` +
	`gAALYYAAC8GAAAwRgAAMcYAADKGAAAzhgAANMYAADXGAAA3BgAAOEYAAD/GAAAIBkAADsZ` +
	`AABVGQAAeBkAAJ0ZAADCGQAA4xkAAPwZAAAPGgAAKxoAAEAaAABWGgAAbxoAAIsaAACfGg` +
	`AA1BoAAPQaAAAgGwAANRsAAFwbAABzGwAAkBsAAL8bAADaGwAABRwAACocAABKHAAAWhwA` +
	`AHQcAACLHAAAqhwAAL4cAADUHAAA6BwAAPwcAAATHQAAOB0AAF0dAACCHQAAqR0AAM4dAA` +
	`DxHQAABx4AABIeAAAqHgAAMx4AAE0eAABYHgAAWx4AAF8eAABkHgAAax4AAHEeAAB1HgAA` +
	`eh4AAIEeAACNHgAAkh4AAJYeAACZHgAAnR4AAKIeAACnHgAAvB4AAMAeAADMHgAA2B4AAO` +
	`QeAADwHgAA/B4AAAgfAAAUHwAAIR8AAC4fAAA+HwAASB8AAGMfAAB7HwAAjR8AAKMfAADK` +
	`HwAA7x8AABkgAAA7IAAAXCAAAHggAACRIAAApCAAALIgAAC8IAAAyyAAANsgAADrIAAA+y` +
	`AAAAshAAAOIQAAFiEAADkhAABNIQAAWyEAAGAhAABxIQAAgiEAAI8hAACdIQAAryEAALgh` +
	`AADGIQAA0SEAANwhAADvIQAA/SEAAAoiAAAfIgAAKCIAADMiAABFIgAAYSIAAHsiAACWIg` +
	`AAryIAALgiAADDIgAAzSIAANgiAADoIgAABiMAABgjAAAgIwAALiMAAEcjAABSIwAAYCMA` +
	`AG8jAAB/IwAAjiMAAJQjAACcIwAAoiMAAK8jAADDIwAA7CMAAPcjAAABJAAAByQAABkkAA` +
	`AxJAAAOyQAAEskAABaJAAAZCQAAHIkAAB3JAAAhiQAAJMkAACiJAAAsCQAAMEkAADPJAAA` +
	`2CQAAOEkAADwJAAA/CQAAAolAAAYJQAAJiUAADUlAAA8JQAAVCUAAGElAABpJQAAcSUAAH` +
	`slAACAJQAAiCUAAKwlAAC6JQAAxyUAANklAADgJQAA5yUAAAoAAAAVAAAAFgAAABcAAAAY` +
	`AAAAGQAAABoAAAAbAAAAHAAAAB0AAAAeAAAAHwAAACAAAAAhAAAAIgAAACMAAAAkAAAAJQ` +
	`AAACYAAAAnAAAAKAAAACkAAAAqAAAAKwAAACwAAAAtAAAALgAAAC8AAAAwAAAAMQAAADIA` +
	`AAAzAAAANAAAADUAAAA2AAAANwAAADgAAAA5AAAAOgAAADsAAAA8AAAAPQAAAD4AAAA/AA` +
	`AARgAAAFEAAABVAAAACgAAAAAAAAAAAAAACwAAAAAAAAAQFwAADAAAAAAAAAAYFwAADQAA` +
	`AAAAAAAkFwAADgAAAAAAAAAsFwAADwAAAAIAAAAAAAAADwAAAAQAAAAAAAAAEAAAAAQAAA` +
	`A4FwAAFAAAAAQAAABAFwAAEgAAAAQAAABIFwAAFAAAAAQAAAAkFwAAFAAAAAQAAABQFwAA` +
	`EwAAAAUAAABYFwAADwAAAAYAAAAAAAAADwAAAAcAAAAAAAAADwAAAAgAAAAAAAAADwAAAA` +
	`oAAAAAAAAADwAAAA0AAAAAAAAADwAAAA4AAAAAAAAAEAAAABIAAAA4FwAADwAAABQAAAAA` +
	`AAAAEAAAABQAAAA4FwAADwAAABYAAAAAAAAADwAAABcAAAAAAAAAEgAAABkAAABgFwAAFA` +
	`AAABkAAABoFwAADwAAAB0AAAAAAAAAEQAAAB4AAAAQFwAAEgAAACEAAABIFwAADwAAACMA` +
	`AAAAAAAAEgAAACMAAABIFwAADwAAACsAAAAAAAAARgAAACwAAAAAAAAARwAAACwAAAA4Fw` +
	`AASAAAACwAAAAQFwAASQAAACwAAABwFwAASgAAACwAAAB8FwAASwAAACwAAACIFwAATAAA` +
	`ACwAAACQFwAASwAAACwAAACYFwAASwAAACwAAACgFwAASwAAACwAAACoFwAASwAAACwAAA` +
	`CwFwAASwAAACwAAAAIFwAASwAAACwAAAAAFwAATgAAACwAAAC4FwAATwAAACwAAADQFwAA` +
	`SwAAACwAAADYFwAASwAAACwAAADgFwAATQAAACwAAADoFwAASwAAACwAAAD4FgAASwAAAC` +
	`wAAABIFwAATwAAACwAAAAkFwAASwAAACwAAAD0FwAASwAAACwAAABgFwAATAAAACwAAAD8` +
	`FwAATwAAACwAAAAEGAAAUAAAACwAAAAMGAAAUQAAAC0AAAAAAAAAUwAAAC0AAAAUGAAAUw` +
	`AAAC0AAAAcGAAAUgAAAC0AAADgFwAAUgAAAC0AAAAkGAAAUgAAAC0AAABgFwAAVAAAAC0A` +
	`AAAsGAAAEgAAAC4AAABIFwAABQAMAKUAAAAHAAAAxAAAAAkAAACeAAAACQAAAMMAAAALAA` +
	`AAQwAAACYAKwDAAAAAJgAAAMgAAAAnACsAwAAAACgAKwDAAAAAKQAqAMEAAAAqACsAwAAA` +
	`ACsAAAAEAAAAKwAAAAUAAAArAAAABgAAACsAAAAHAAAAKwAAAEAAAAArAAAAQgAAACsAAA` +
	`BEAAAAKwArAJUAAAArAC0AmQAAACsAGQCkAAAAAQAgAAMAAAABACcAqAAAAAEAKQCpAAAA` +
	`BAAzAAMAAAAEAAkAXwAAAAQABwBhAAAABAAIAHAAAAAEAAUAfgAAAAQAEACAAAAABAAKAK` +
	`wAAAAEAAsArAAAAAQACQC2AAAABgAMAHwAAAAIAA4AfwAAAAkAIAADAAAACQAAAJYAAAAJ` +
	`AAAAygAAAAoAHQDCAAAADAAeAIkAAAAOAAAAnwAAABAAAwB1AAAAEAAEAHUAAAASAAEAeg` +
	`AAABIAEwCgAAAAFAAsAGIAAAAUAAAAggAAABQAFACGAAAAFAAXAIcAAAAUAAAAkQAAABQA` +
	`EQCTAAAAFAAoAJQAAAAWABQAgQAAABcAAACLAAAAFwAAAIwAAAAXAAAAjQAAABcAAACOAA` +
	`AAGAA7AJgAAAAYADwAvAAAABkAJQADAAAAGQArAGMAAAAZACAAbgAAABkAEgCPAAAAGQA6` +
	`AK0AAAAZACEAsQAAABkAIQCyAAAAGQAvALMAAAAZACEAtAAAABkAMAC1AAAAGQAhALcAAA` +
	`AaACIAAwAAAB0AHQB7AAAAHgAbAL8AAAAeAB0AwgAAACEAIAADAAAAIwA9AG8AAAAjAD4A` +
	`dgAAACMAQQC9AAAAJAAzAKEAAAAmADcAAwAAACYAIACuAAAAJwA2AAMAAAAnACAArgAAAC` +
	`gANgADAAAAKAAtAKoAAAApADUAAwAAACkAKgBkAAAAKQAxAG0AAAApADEAqwAAACoANgAD` +
	`AAAAKgAgAK4AAAArACAAAwAAACsAGABXAAAAKwAZAFgAAAArAD8AWQAAACsAQABaAAAAKw` +
	`AfAFsAAAArADgAXAAAACsANgBdAAAAKwAuAGAAAAArACAAcQAAACsAMwByAAAAKwA0AHMA` +
	`AAArACEAdAAAACsAMwB4AAAAKwAVAHkAAAArABoAfQAAACsABgCDAAAAKwANAIQAAAArAA` +
	`8AhQAAACsAAgCIAAAAKwAcAIoAAAArAB0AkAAAACsAFgCSAAAAKwAgAJcAAAArACMAmwAA` +
	`ACsAIACcAAAAKwAzAJ0AAAArACAAoAAAACsAJACnAAAAKwAnAKgAAAArACkAqQAAACsAMg` +
	`CvAAAAKwA5ALAAAAArACAAuAAAACsAMwC5AAAAKwA0ALoAAAArACEAuwAAACsAJgC+AAAA` +
	`KwAgAMYAAAArACcAxwAAACYAAAAAAAAAIQAAAPgWAAAJAAAAqBYAAMInAAAAAAAAJwAAAA` +
	`AAAAAhAAAA+BYAAAkAAAC4FgAA1icAAAAAAAAoAAAAAAAAACEAAAAAFwAACQAAAMgWAADn` +
	`JwAAAAAAACkAAAAAAAAAIQAAAAgXAAAJAAAA2BYAAPgnAAAAAAAAKgAAAAAAAAAhAAAA+B` +
	`YAAAkAAADoFgAAESgAAAAAAAArAAAAAQAAAAEAAAAAAAAACQAAAAAAAAAiKAAAsScAAAIA` +
	`AACFJwAAjCcAAAIAAACVJwAAjCcAAAIAAACcJwAAjCcAAAIAAACjJwAAjCcAAAIAAACqJw` +
	`AAjCcAAAMAAwABAAAA6iUAAAgAAABbAQUAWQIGAHAQNQAAAA4ABgABAAMAAADxJQAAmAAA` +
	`ABUCAEASYRIEFQAIAFJTBgArA4QAAAABIRoCCAAaA8UAcSAUADIAVFIFAHEQRwACAAwCbi` +
	`ArABIAVFEFAHEQRwABAAwBbiAsAAEAVFAFABIRcSBKABAAVFAFAHEQRwAAAAwAGgECAG4g` +
	`LwAQAFRQBQBxEEcAAAAMAFRRBQBxEEcAAQAMAW4QKQABAAwBchATAAEACgFuIC4AEABUUA` +
	`UAcSBKAEAAVFAFAHEQRwAAAAwAbiAwAEAAVFAFAHEQRwAAAAwAbhAoAAAAVFAFAHEQRwAA` +
	`AAwAbhAqAAAAVFAFABoBmgBuIFoAEAAMAB8AGABUUQUAcRBHAAEADAFuMCUAEAQOAAEhKJ` +
	`EUAAIACAAojRQAkAAIACiCAAAAAQQAAAAAAHkAAAALAAAAewAAAH8AAAACAAIAAQAAABIm` +
	`AAAGAAAAWwEHAHAQNQAAAA4AAwABAAIAAAAZJgAADAAAAFQgBwBxEEcAAAAMABMBCABuID` +
	`AAEAAOAAIAAgABAAAAICYAAAYAAABbAQgAcBA1AAAADgALAAoAAQAAACcmAAAGAAAAVBAI` +
	`AG4QbAAAAA4AAgACAAEAAAA3JgAABgAAAFsBCQBwEDUAAAAOAAQAAgACAAAAPiYAAD8AAA` +
	`ASEXIQEwADAAoANRA5AFQgCQBUAAoAcSBKABAAVCAJAFQACgBxEEcAAAAMABoBAgBuIC8A` +
	`EABUIAkAVAAKAHEQRwAAAAwAVCEJAFQRCgBxEEcAAQAMAW4QKQABAAwBchATAAEACgFuIC` +
	`4AEABUIAkAVAAKABIBcSBKABAADgAAAAcABQABAAAATCYAABoAAABUIAkAVAAKAHEQSQAA` +
	`AAoAOAADAA4APQX//xIANVD8/1QhCQBUEQoAcRBNAAEA2AAAASj1BwAFAAMAAABdJgAAHw` +
	`AAAFQgCQBUAAoAcRBJAAAACgA4AAMADgA9Bv//VCAJAFQACgCQAQQGcjAzAEMBDAFyEDQA` +
	`AQAMAXEgTAAQACjsAAACAAIAAQAAAGomAAAGAAAAWwEKAHAQNQAAAA4ABQABAAMAAABxJg` +
	`AAbwAAABLjVEAKACIBGQBxAEsAAAAMAnAgJgAhAHEgSAAQAFRACgBxEEcAAAAMABMBCABu` +
	`IDAAEABUQAoAcRBHAAAADAAVAQgAbiAsABAAIgAaAHAwMQAwA1RBCgBxEEcAAQAMAW4gLQ` +
	`ABAFRBCgBUQgoAcRBHAAIADAJuME4AIQBUQAoAcRBHAAAADAAaAQIAbiAvABAAVEAKAHEQ` +
	`RwAAAAwAVEEKAHEQRwABAAwBbhApAAEADAFyEBMAAQAKAW4gLgAQAFRACgBxEEcAAAAMAC` +
	`IBKQBwIEAAQQBuICcAEAAOAAAAAgABAAEAAACEJgAACQAAAHAQAAABABIAXBATAGkBEgAO` +
	`AAAAAgABAAAAAACMJgAAAwAAAFQQFAARAAAAAgACAAAAAACSJgAAAwAAAFsBFAARAQAAAg` +
	`ABAAAAAACZJgAAAwAAAFUQEwAPAAAAAgACAAAAAACfJgAAAwAAAFwBEwAPAQAAAQAAAAAA` +
	`AACmJgAAAwAAAGIAEgARAAAAAgACAAIAAACrJgAABAAAAHAgYAAQAA4AAQABAAEAAACyJg` +
	`AABAAAAHAQXwAAAA4ABwADAAMAAQC4JgAAGQAAABLwcRAXAAQADAFuMBYAUQYKATkBAwAP` +
	`AAEQKP4NARoCCAAaA3cAcTAVADIBKPUNASjzAAABAAAABwABAAECERcfDgAAAQAAAAEAAA` +
	`DJJgAABgAAAGIAEgBuEE8AAAAOAAQAAQADAAEAzyYAADMAAABuEFcAAwAMAG4QVgADAAwB` +
	`bhAHAAEADAETAoAAbjAMABACDABUAQAAOQEKABoACAAaAaMAcSAUABAADgBUAAAAGgFlAG` +
	`4gEgAQAAwAcRA5AAAAKPQNABoBCAAaAqIAcTAVACEAKOsAAAAAAAApAAEAAQEfKgIAAQAC` +
	`AAAA4CYAAAkAAAAiACoAcCBEABAAbiBlAAEADgAAAAIAAQACAAAA6SYAAAYAAABiABIAbi` +
	`BQABAADgADAAIAAwAAAPEmAAAGAAAAYgASAG4wUQAQAg4AAgABAAIAAAD6JgAABgAAAGIA` +
	`EgBuIFIAEAAOAAQAAQADAAAAAScAACQAAAAaAJoAbiBaAAMADAAfABgAFAECAAIBbiBUAB` +
	`MADAFuEBoAAQAMAW4QHQABAAwBEgJuMCQAEAIiACcAcCA8ADAAbiBlAAMADgAGAAIAAwAA` +
	`AAonAABXAAAAEhMiAAQAGgFnAHAgAwAQABoBbABuIDcAUQAKATgBHABgAQQAEwIVADQhFg` +
	`AiAAQAGgFoAHAgAwAQAG4gBQAwABoBQQBxIAYAEAAMAG4wawAEAw4AGgHLAG4gNgAVAAoB` +
	`OAEeAGABBAATAhMANCEYABoBAQBuIAsAEAAaAWoAGgJWAG4gOAAlAAwCbjAKABACGgFpAG` +
	`4gBAAQACjTbiALAFAAGgFpAG4gBAAQACjKAAAGAAMAAwAAAB4nAAA+AAAAIgAEABoBZgBw` +
	`IAMAEAAaAcsAbiA2ABQACgE4AS0AYAEEABMCEwA0IScAGgEBAG4gCwAQABoBagAaAlYAbi` +
	`A4ACQADAJuMAoAEAIaAWsAbjAJABAFGgFpAG4gBAAQABoBRQBxIAYAEAAMABIhbjBrAAMB` +
	`DgBuIAsAQAAo6AMAAgADAAAAMCcAAAkAAAAiACYAcDA6ABACbiBlAAEADgAAAAIAAQABAA` +
	`AAOScAAAkAAABuEFUAAQAMAG4QMgAAAAwAEQAAAAUABAACAAAAPicAABwAAAASEDICBgAS` +
	`IDICAwAOABLwMgMIABoAAABwIFMAAQAo924QCAAEAAwAbhARAAAADABwIFMAAQAo6wIAAg` +
	`ACAAAAUCcAAAcAAABvIAEAEABuIG0AEAAOAAAABAACAAIAAABZJwAAKAAAAHAQYQACAG8g` +
	`AgAyAHAQZwACAG4QWAACAAwAbhANAAAADABuIG0AAgAUAAIAAgFuIFQAAgAMAG4QGgAAAA` +
	`wAIgEoAHAgPgAhAG4gGAAQAA4ABwABAAUAAQBmJwAAYAAAAG4QXAAGAAwAbhAfAAAADABu` +
	`EBsAAAAMADkAAwAOAG4QIwAAAAoBbhAgAAAACgJuECEAAAAKA24QIgAAAAoAcFBeABYyKO` +
	`wNACIACQBwEA4AAABuEFwABgAMAW4QHwABAAwBbiAeAAEAFAECAAIBbiBUABYADAFuEBoA` +
	`AQAMAVICAwBuEBkAAQAKA24QDwAAAAoEsUNSBAMAsUNSBAIAbhAcAAEACgFuEBAAAAAKBb` +
	`FRUgACAJEAAQBwUF4AJkMorwAAAAAiAAEAAQEgIwQAAgACAAAAeycAAA8AAABSMAEA3QAA` +
	`MBMBIAAzEAcAEhBwIGYAAgAOABIAKPsAAEwMAAAAAAAAAAAAAAAAAABYDAAAAAAAAAAAAA` +
	`AAAAAAZAwAAAAAAAAAAAAAAAAAAHAMAAAAAAAAAAAAAAAAAAB8DAAAAAAAAAAAAAAAAAAA` +
	`AQAAACIAAAABAAAAEwAAAAEAAAAPAAAAAgAAAAAAAAADAAAAAAAAAAAAAAACAAAAIwAjAA` +
	`MAAAAjACMAJQAAAAEAAAAAAAAAAgAAAAQAHgABAAAAIwAAAAIAAAAjAC4AAgAAAAIAAAAB` +
	`AAAAKwAAAAIAAAArABkABAAAAAAAAAAAAAAAAwAAAAAAAAAEAAAAAQAAAAMAAAACAAAABA` +
	`AAAAEAAAAHAAAAAQAAAAkAAAABAAAADAAAAAEAAAAOAAAACQAAABQAAAAAAAAAAAAAAAAA` +
	`AAAAAAAAAgAAABQAFQABAAAAFQAAAAEAAAAeAAAABAAAAB4AAAAAAAAAAQAAACoAAAACAA` +
	`AAKwAAAAIAAAArACMAAQAAAC0AAAACAAAADQAAAAIAAAAUAAAAAQAAACEAAAACAAAAKwAt` +
	`AAAAAyovKgABMAAGPGluaXQ+ABJERUZBVUxUX0lOUFVUX1RZUEUAFURFRkFVTFRfS0VZQk` +
	`9BUkRfQ09ERQAORklMRV9PUEVOX0NPREUADkZJTEVfU0FWRV9DT0RFAARGeW5lABVHb05h` +
	`dGl2ZUFjdGl2aXR5LmphdmEAAUkAA0lJSQAESUlJSQADSUxMAARJTExMAAFMAAJMSQADTE` +
	`lJAAJMTAADTExJAANMTEwAHExhbmRyb2lkL2FwcC9OYXRpdmVBY3Rpdml0eTsAH0xhbmRy` +
	`b2lkL2NvbnRlbnQvQ29tcG9uZW50TmFtZTsAGUxhbmRyb2lkL2NvbnRlbnQvQ29udGV4dD` +
	`sAGExhbmRyb2lkL2NvbnRlbnQvSW50ZW50OwAhTGFuZHJvaWQvY29udGVudC9wbS9BY3Rp` +
	`dml0eUluZm87ACNMYW5kcm9pZC9jb250ZW50L3BtL1BhY2thZ2VNYW5hZ2VyOwAjTGFuZH` +
	`JvaWQvY29udGVudC9yZXMvQ29uZmlndXJhdGlvbjsAH0xhbmRyb2lkL2NvbnRlbnQvcmVz` +
	`L1Jlc291cmNlczsAF0xhbmRyb2lkL2dyYXBoaWNzL1JlY3Q7ABFMYW5kcm9pZC9uZXQvVX` +
	`JpOwAaTGFuZHJvaWQvb3MvQnVpbGQkVkVSU0lPTjsAE0xhbmRyb2lkL29zL0J1bmRsZTsA` +
	`FExhbmRyb2lkL29zL0lCaW5kZXI7ABdMYW5kcm9pZC90ZXh0L0VkaXRhYmxlOwAaTGFuZH` +
	`JvaWQvdGV4dC9UZXh0V2F0Y2hlcjsAEkxhbmRyb2lkL3V0aWwvTG9nOwAzTGFuZHJvaWQv` +
	`dmlldy9LZXlDaGFyYWN0ZXJNYXAkVW5hdmFpbGFibGVFeGNlcHRpb247AB5MYW5kcm9pZC` +
	`92aWV3L0tleUNoYXJhY3Rlck1hcDsAKkxhbmRyb2lkL3ZpZXcvVmlldyRPbkxheW91dENo` +
	`YW5nZUxpc3RlbmVyOwATTGFuZHJvaWQvdmlldy9WaWV3OwAlTGFuZHJvaWQvdmlldy9WaW` +
	`V3R3JvdXAkTGF5b3V0UGFyYW1zOwAVTGFuZHJvaWQvdmlldy9XaW5kb3c7ABtMYW5kcm9p` +
	`ZC92aWV3L1dpbmRvd0luc2V0czsALUxhbmRyb2lkL3ZpZXcvaW5wdXRtZXRob2QvSW5wdX` +
	`RNZXRob2RNYW5hZ2VyOwAZTGFuZHJvaWQvd2lkZ2V0L0VkaXRUZXh0OwApTGFuZHJvaWQv` +
	`d2lkZ2V0L0ZyYW1lTGF5b3V0JExheW91dFBhcmFtczsAI0xkYWx2aWsvYW5ub3RhdGlvbi` +
	`9FbmNsb3NpbmdNZXRob2Q7AB5MZGFsdmlrL2Fubm90YXRpb24vSW5uZXJDbGFzczsADkxq` +
	`YXZhL2lvL0ZpbGU7ABhMamF2YS9sYW5nL0NoYXJTZXF1ZW5jZTsAFUxqYXZhL2xhbmcvRX` +
	`hjZXB0aW9uOwAdTGphdmEvbGFuZy9Ob1N1Y2hNZXRob2RFcnJvcjsAEkxqYXZhL2xhbmcv` +
	`T2JqZWN0OwAUTGphdmEvbGFuZy9SdW5uYWJsZTsAEkxqYXZhL2xhbmcvU3RyaW5nOwASTG` +
	`phdmEvbGFuZy9TeXN0ZW07ABVMamF2YS9sYW5nL1Rocm93YWJsZTsAI0xvcmcvZ29sYW5n` +
	`L2FwcC9Hb05hdGl2ZUFjdGl2aXR5JDE7ACNMb3JnL2dvbGFuZy9hcHAvR29OYXRpdmVBY3` +
	`Rpdml0eSQyOwAjTG9yZy9nb2xhbmcvYXBwL0dvTmF0aXZlQWN0aXZpdHkkMzsAJUxvcmcv` +
	`Z29sYW5nL2FwcC9Hb05hdGl2ZUFjdGl2aXR5JDQkMTsAI0xvcmcvZ29sYW5nL2FwcC9Hb0` +
	`5hdGl2ZUFjdGl2aXR5JDQ7ACFMb3JnL2dvbGFuZy9hcHAvR29OYXRpdmVBY3Rpdml0eTsA` +
	`FE5VTUJFUl9LRVlCT0FSRF9DT0RFAAlPcGVuIEZpbGUAFlBBU1NXT1JEX0tFWUJPQVJEX0` +
	`NPREUAB1NES19JTlQAGFNJTkdMRUxJTkVfS0VZQk9BUkRfQ09ERQAJU2F2ZSBGaWxlAAFW` +
	`AAJWSQADVklJAAVWSUlJSQAEVklJTAACVkwAA1ZMSQAFVkxJSUkAClZMSUlJSUlJSUkAA1` +
	`ZMTAACVloAAVoAAlpMAANaTEkAA1pMWgATW0xqYXZhL2xhbmcvU3RyaW5nOwACXHwACmFj` +
	`Y2VzcyQwMDAACmFjY2VzcyQwMDIACmFjY2VzcyQxMDAACmFjY2VzcyQxMDIACmFjY2Vzcy` +
	`QyMDAACmFjY2VzcyQzMDAACmFjY2VzcyQ0MDAAC2FjY2Vzc0ZsYWdzAAthZGRDYXRlZ29y` +
	`eQAOYWRkQ29udGVudFZpZXcACGFkZEZsYWdzABlhZGRPbkxheW91dENoYW5nZUxpc3Rlbm` +
	`VyABZhZGRUZXh0Q2hhbmdlZExpc3RlbmVyABBhZnRlclRleHRDaGFuZ2VkABRhbmRyb2lk` +
	`LmFwcC5saWJfbmFtZQAlYW5kcm9pZC5pbnRlbnQuYWN0aW9uLkNSRUFURV9ET0NVTUVOVA` +
	`AjYW5kcm9pZC5pbnRlbnQuYWN0aW9uLk9QRU5fRE9DVU1FTlQAKGFuZHJvaWQuaW50ZW50` +
	`LmFjdGlvbi5PUEVOX0RPQ1VNRU5UX1RSRUUAIGFuZHJvaWQuaW50ZW50LmNhdGVnb3J5Lk` +
	`9QRU5BQkxFAB9hbmRyb2lkLmludGVudC5leHRyYS5NSU1FX1RZUEVTABphbmRyb2lkLmlu` +
	`dGVudC5leHRyYS5USVRMRQAXYXBwbGljYXRpb24veC1kaXJlY3RvcnkAEWJlZm9yZVRleH` +
	`RDaGFuZ2VkAAxicmluZ1RvRnJvbnQACGNvbnRhaW5zAA1jcmVhdGVDaG9vc2VyAA5kb0hp` +
	`ZGVLZXlib2FyZAAOZG9TaG93RmlsZU9wZW4ADmRvU2hvd0ZpbGVTYXZlAA5kb1Nob3dLZX` +
	`lib2FyZAABZQAGZXF1YWxzACFleGNlcHRpb24gcmVhZGluZyBLZXlDaGFyYWN0ZXJNYXAA` +
	`EmZpbGVQaWNrZXJSZXR1cm5lZAAMZmluZFZpZXdCeUlkAANnZXQAD2dldEFic29sdXRlUG` +
	`F0aAAPZ2V0QWN0aXZpdHlJbmZvAAtnZXRDYWNoZURpcgAMZ2V0Q29tcG9uZW50ABBnZXRD` +
	`b25maWd1cmF0aW9uAAdnZXREYXRhAAxnZXREZWNvclZpZXcACWdldEhlaWdodAAJZ2V0SW` +
	`50ZW50ABFnZXRQYWNrYWdlTWFuYWdlcgAMZ2V0UmVzb3VyY2VzAAtnZXRSb290VmlldwAT` +
	`Z2V0Um9vdFdpbmRvd0luc2V0cwAHZ2V0UnVuZQAJZ2V0U3RyaW5nABBnZXRTeXN0ZW1TZX` +
	`J2aWNlABpnZXRTeXN0ZW1XaW5kb3dJbnNldEJvdHRvbQAYZ2V0U3lzdGVtV2luZG93SW5z` +
	`ZXRMZWZ0ABlnZXRTeXN0ZW1XaW5kb3dJbnNldFJpZ2h0ABdnZXRTeXN0ZW1XaW5kb3dJbn` +
	`NldFRvcAAHZ2V0VGV4dAAJZ2V0VG1wZGlyAAhnZXRXaWR0aAAJZ2V0V2luZG93AA5nZXRX` +
	`aW5kb3dUb2tlbgAcZ2V0V2luZG93VmlzaWJsZURpc3BsYXlGcmFtZQAQZ29OYXRpdmVBY3` +
	`Rpdml0eQAGaGVpZ2h0AAxoaWRlS2V5Ym9hcmQAF2hpZGVTb2Z0SW5wdXRGcm9tV2luZG93` +
	`AAlpZ25vcmVLZXkADGlucHV0X21ldGhvZAANaW5zZXRzQ2hhbmdlZAAOa2V5Ym9hcmREZW` +
	`xldGUADWtleWJvYXJkVHlwZWQABGxlZnQABmxlbmd0aAAEbG9hZAALbG9hZExpYnJhcnkA` +
	`EmxvYWRMaWJyYXJ5IGZhaWxlZAAnbG9hZExpYnJhcnk6IG5vIG1hbmlmZXN0IG1ldGFkYX` +
	`RhIGZvdW5kAAltVGV4dEVkaXQACG1ldGFEYXRhAARuYW1lABBvbkFjdGl2aXR5UmVzdWx0` +
	`ABZvbkNvbmZpZ3VyYXRpb25DaGFuZ2VkAAhvbkNyZWF0ZQAOb25MYXlvdXRDaGFuZ2UADW` +
	`9uVGV4dENoYW5nZWQACHB1dEV4dHJhAAxyZXF1ZXN0Rm9jdXMAA3J1bgANcnVuT25VaVRo` +
	`cmVhZAALc2V0RGFya01vZGUADXNldEltZU9wdGlvbnMADHNldElucHV0VHlwZQAPc2V0TG` +
	`F5b3V0UGFyYW1zAAxzZXRTZWxlY3Rpb24AB3NldFRleHQAB3NldFR5cGUADXNldFZpc2li` +
	`aWxpdHkACnNldHVwRW50cnkADHNob3dGaWxlT3BlbgAMc2hvd0ZpbGVTYXZlAAxzaG93S2` +
	`V5Ym9hcmQADXNob3dTb2Z0SW5wdXQABXNwbGl0ABZzdGFydEFjdGl2aXR5Rm9yUmVzdWx0` +
	`AAtzdWJTZXF1ZW5jZQAGdGhpcyQwAAZ0aGlzJDEACHRvU3RyaW5nAAN0b3AABnVpTW9kZQ` +
	`AidW5rbm93biBrZXlib2FyZCB0eXBlLCB1c2UgZGVmYXVsdAAMdXBkYXRlTGF5b3V0AAt1` +
	`cGRhdGVUaGVtZQAQdmFsJGtleWJvYXJkVHlwZQAFdmFsdWUABXdpZHRoAAF8AFECAAAHDg` +
	`BUAAdKDy0CD2h5lphptAEXD1uWlpellgJjLCM8IAJzSgCDAQEABw4AhgEABw60AN0BAQAH` +
	`DgDgAQkAAAAAAAAAAAAHDloA9gEBAAcOAJECAQAHHWl40gEbD4kAgwIEAAAAAAcOrQJ6HS` +
	`09dQD5AQQAAAAABw6qGi0A5gEBAAcOAOkBAAcd4bS1W5a2tAEXEAIk4AAxAAcOOD8tABsB` +
	`AAcOABsCAAAHDgAbAQAHDgAbAgAABw4AGwAHDgAbAgAABw4AGwEABw4AsQEDAAAABx2HNA` +
	`J7LCAegwB7AAcOWgDIAQAHDkujTEt/Ansdh0seAOYBAAcOAjaGAIwBAQAHDloAoAECAAAH` +
	`DloATQEABw5aAH8ABw6HtIiMAJABAQAHHXjheESWAncd4Vq0ajwApAECAAAHDnjhWrdaWq` +
	`UCex0AUQEABw4CJ4YANgAHDgChAgMAAAAHDgIMaAJ5HTxsSwCxAgEABw48PADXAQEABw48` +
	`PDy1tIwAOwAHDsMCDiwCdh2HhUweWrW0/9AAtgIBAAcOljwbAAIbAckBGlICHAJeBACmAR` +
	`4CGwHJARpPAhsByQEaZAIbAckBGkUCGwHJARpnB0QAAAgEAAQBBAIEAgQDBAEAAgEBBZAg` +
	`AZAgOoCABIgZOwGoGQABAQEHkCA8gIAE6Bs9AYQcAAEBAQiQID6AgASsHD8ByBwAAQEDCZ` +
	`AgQICABOQcQQGAHQEBkB4BAdQeAAEBAQqQIESAgASkH0UBwB8IAhQKCxoBGgEaARoBGgEa` +
	`ARoBChMCAQJGgYAEsCEBiCDUIQGIIOwhAYgghCIBiCCcIgGIILQiAYggzCIBiCDkIgaCAg` +
	`AGCPwiBAjQIwGCAgABggIAAYICAAEC7CMFggIAAQLwJAEIlCUBCLAlAQjMJU8A6CUBAMAm` +
	`AQCAKAEAjCkJALApBwTUKQEBnCoBAbwqCACcKwEE+CwAEQAAAAAAAAABAAAAAAAAAAEAAA` +
	`DMAAAAcAAAAAIAAAAvAAAAoAMAAAMAAABCAAAAXAQAAAQAAAAVAAAAdAcAAAUAAABuAAAA` +
	`HAgAAAYAAAAGAAAAjAsAAAMQAAAFAAAATAwAAAEgAAAlAAAAiAwAAAYgAAAFAAAAqBYAAA` +
	`EQAAAjAAAA+BYAAAIgAADMAAAANBgAAAMgAAAlAAAA6iUAAAQgAAAGAAAAhScAAAUgAAAB` +
	`AAAAsScAAAAgAAAGAAAAwicAAAAQAAABAAAAvCgAAA==` +
	``

// signPKCS7 does the minimal amount of work necessary to embed an RSA
// signature into a PKCS#7 certificate.
//
// We prepare the certificate using the x509 package, read it back in
// to our custom data type and then write it back out with the signature.
func signPKCS7(rand io.Reader, priv *rsa.PrivateKey, msg []byte) ([]byte, error) {
	const serialNumber = 0x5462c4dd // arbitrary
	name := pkix.Name{CommonName: "gomobile"}

	template := &x509.Certificate{
		SerialNumber:       big.NewInt(serialNumber),
		SignatureAlgorithm: x509.SHA1WithRSA,
		Subject:            name,
		NotAfter:           time.Date(2120, time.January, 1, 12, 0, 0, 0, time.UTC), // more than 50 years for Google Play Store
	}

	b, err := x509.CreateCertificate(rand, template, template, priv.Public(), priv)
	if err != nil {
		return nil, err
	}

	c := certificate{}
	if _, err := asn1.Unmarshal(b, &c); err != nil {
		return nil, err
	}

	h := sha1.New()
	_, err = h.Write(msg)
	if err != nil {
		return nil, err
	}
	hashed := h.Sum(nil)

	signed, err := rsa.SignPKCS1v15(rand, priv, crypto.SHA1, hashed)
	if err != nil {
		return nil, err
	}

	content := pkcs7SignedData{
		ContentType: oidSignedData,
		Content: signedData{
			Version: 1,
			DigestAlgorithms: []pkix.AlgorithmIdentifier{{
				Algorithm:  oidSHA1,
				Parameters: asn1.RawValue{Tag: 5},
			}},
			ContentInfo:  contentInfo{Type: oidData},
			Certificates: c,
			SignerInfos: []signerInfo{{
				Version: 1,
				IssuerAndSerialNumber: issuerAndSerialNumber{
					Issuer:       name.ToRDNSequence(),
					SerialNumber: serialNumber,
				},
				DigestAlgorithm: pkix.AlgorithmIdentifier{
					Algorithm:  oidSHA1,
					Parameters: asn1.RawValue{Tag: 5},
				},
				DigestEncryptionAlgorithm: pkix.AlgorithmIdentifier{
					Algorithm:  oidRSAEncryption,
					Parameters: asn1.RawValue{Tag: 5},
				},
				EncryptedDigest: signed,
			}},
		},
	}

	return asn1.Marshal(content)
}

type pkcs7SignedData struct {
	ContentType asn1.ObjectIdentifier
	Content     signedData `asn1:"tag:0,explicit"`
}

// signedData is defined in rfc2315, section 9.1.
type signedData struct {
	Version          int
	DigestAlgorithms []pkix.AlgorithmIdentifier `asn1:"set"`
	ContentInfo      contentInfo
	Certificates     certificate  `asn1:"tag:0,explicit"`
	SignerInfos      []signerInfo `asn1:"set"`
}

type contentInfo struct {
	Type asn1.ObjectIdentifier
	// Content is optional in PKCS#7 and not provided here.
}

// certificate is defined in rfc2459, section 4.1.
type certificate struct {
	TBSCertificate     tbsCertificate
	SignatureAlgorithm pkix.AlgorithmIdentifier
	SignatureValue     asn1.BitString
}

// tbsCertificate is defined in rfc2459, section 4.1.
type tbsCertificate struct {
	Version      int `asn1:"tag:0,default:2,explicit"`
	SerialNumber int
	Signature    pkix.AlgorithmIdentifier
	Issuer       pkix.RDNSequence // pkix.Name
	Validity     validity
	Subject      pkix.RDNSequence // pkix.Name
	SubjectPKI   subjectPublicKeyInfo
}

// validity is defined in rfc2459, section 4.1.
type validity struct {
	NotBefore time.Time
	NotAfter  time.Time
}

// subjectPublicKeyInfo is defined in rfc2459, section 4.1.
type subjectPublicKeyInfo struct {
	Algorithm        pkix.AlgorithmIdentifier
	SubjectPublicKey asn1.BitString
}

type signerInfo struct {
	Version                   int
	IssuerAndSerialNumber     issuerAndSerialNumber
	DigestAlgorithm           pkix.AlgorithmIdentifier
	DigestEncryptionAlgorithm pkix.AlgorithmIdentifier
	EncryptedDigest           []byte
}

type issuerAndSerialNumber struct {
	Issuer       pkix.RDNSequence // pkix.Name
	SerialNumber int
}

// Various ASN.1 Object Identifies, mostly from rfc3852.
var (
	oidData          = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 7, 1}
	oidSignedData    = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 7, 2}
	oidSHA1          = asn1.ObjectIdentifier{1, 3, 14, 3, 2, 26}
	oidRSAEncryption = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 1}
)
