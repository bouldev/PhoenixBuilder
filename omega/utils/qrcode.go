package utils

//
//import (
//	qrcode "github.com/skip2/go-qrcode"
//)
//
//const (
//	QUALITY_LOW      = qrcode.Low
//	QUALITY_MEDIUM   = qrcode.Medium
//	QUALITY_HIGH     = qrcode.High
//	QUALITY_HEIGHEST = qrcode.Highest
//)
//
//type ColorCollector interface {
//	NewRow()
//	BlackBlock()
//	WhiteBlock()
//}
//
//func QRCode(content string, collector ColorCollector, level qrcode.RecoveryLevel) error {
//	qr, err := qrcode.New(content, level)
//	if err != nil {
//		return err
//	}
//
//	for ir, row := range qr.Bitmap() {
//		lr := len(row)
//		if ir == 0 || ir == 1 || ir == 2 ||
//			ir == lr-1 || ir == lr-2 || ir == lr-3 {
//			continue
//		}
//		for ic, col := range row {
//			lc := len(qr.Bitmap())
//			if ic == 0 || ic == 1 || ic == 2 ||
//				ic == lc-1 || ic == lc-2 || ic == lc-3 {
//				continue
//			}
//			if col {
//				collector.WhiteBlock()
//			} else {
//				collector.BlackBlock()
//			}
//		}
//		collector.NewRow()
//	}
//	return nil
//}
//
//type BasicColorCollector struct {
//	fgColor string
//	bgColor string
//	QrImage []string
//	counter int
//}
//
//type MCColorCollector struct {
//	fgColor string
//	bgColor string
//	QrImage []string
//	counter int
//}
//
//func NewTerminalColorCollector() *BasicColorCollector {
//	return &BasicColorCollector{
//		bgColor: "\033[38;5;0m██\033[0m",
//		fgColor: "\033[48;5;7m  \033[0m",
//		QrImage: []string{},
//		counter: -1,
//	}
//}
//
//func NewMCColorCollector() *BasicColorCollector {
//	return &BasicColorCollector{
//		bgColor: "§0██",
//		fgColor: "§f██",
//		QrImage: []string{},
//		counter: -1,
//	}
//}
//
//func (t *BasicColorCollector) NewRow() {
//	t.QrImage = append(t.QrImage, "")
//	t.counter++
//}
//
//func (t *BasicColorCollector) BlackBlock() {
//	if len(t.QrImage) == 0 {
//		t.NewRow()
//	}
//	t.QrImage[t.counter] = t.QrImage[t.counter] + t.fgColor
//}
//
//func (t *BasicColorCollector) WhiteBlock() {
//	if len(t.QrImage) == 0 {
//		t.NewRow()
//	}
//	t.QrImage[t.counter] = t.QrImage[t.counter] + t.bgColor
//}
//
//func TerminalQrCode(content string) []string {
//	tc := NewTerminalColorCollector()
//	if err := QRCode(content, tc, QUALITY_LOW); err != nil {
//		return nil
//	}
//	return tc.QrImage
//}
//
//func MCQrCode(content string) []string {
//	tc := NewMCColorCollector()
//	if err := QRCode(content, tc, QUALITY_LOW); err != nil {
//		return nil
//	}
//	return tc.QrImage
//}
