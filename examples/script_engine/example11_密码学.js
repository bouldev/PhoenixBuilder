// 我们在fb的js解释器中内置了 Crypto.js 的部分库
// 包括 aes,md5,rc4,sha256,tripledes,hmac-md5,hmac-256

// 以下是以 aes 库为例的一个演示

const kPassphrase = "pass";
const ivStr = 'v8.gamma.crypto'
let pass = '这是原始字符串'

let key = CryptoJS.enc.Utf8.parse(kPassphrase)
let iv = CryptoJS.enc.Utf8.parse(ivStr)

let c = CryptoJS.AES.encrypt(pass, key, {
    iv: iv,
}).ciphertext.toString(CryptoJS.enc.Base64)

engine.message(c)

// 解密
let cipherParams = CryptoJS.lib.CipherParams.create({
    ciphertext: CryptoJS.enc.Base64.parse(c),
})
let result = CryptoJS.AES.decrypt(cipherParams, key, {
    iv: iv,
})
engine.message(result.toString(CryptoJS.enc.Utf8))