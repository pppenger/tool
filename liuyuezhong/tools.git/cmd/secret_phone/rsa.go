package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"log"
)

var privateKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEqAIBAAKCAQEAmjZzbDPaO9tA7qMMoIZLE7JZDpaMe9mU3a1aATz+Mmi1PMzf
mdOQQmsKipIkN05i9JwBP3bYnNQCHD6L/zsDQOcNrxVdyta1jn+Z2LqfOjCD9tm7
nN36BlppShOnloKBO5459Eq6tT57djMxdbzeB57YDuM4bA1FUNljSRSmGM3ecm+w
PxNc3jatG4PVt54fBWrWb52xL7nAj6oBOe1g14xrdv70e4ruTkI6U8cfXMm6GYR7
KwKGxGoHznBGalkYJXn5QPpKISDLIqwNYl/ikO0qoJvFlg7QAORXDGWK+RIhEPi3
lSLGff5fZqqhtOOoW5vs0dFrAa7N7RkPHXvE/QIDAQABAoIBACY8I+gK/yviE7pX
MNVIFqs+h/gm+ZPngZJo3az662eiMHVvsmzYWKcvFQEAdPxFciGF4IVUuSZBJnjM
RTe18PuRjgIAAS4+h+tZ1dI3iy0JRF7a4zpkiw4yMnLMZrvZhaM7etEICjzPzhqj
XLr9ZetrxdQDIEUiEPb+pXngaE2b1P2i/7xj31OF8UAura/E83lEHOSBgA8TmbsI
1315krHqjjR6s1/o6TQkEVIe0YSc6SkDa7/40g0Ip8yqmXEIoEylOnY/4mBrf6Tn
Htr0xLqL8o2NEzWXCyYlYQjNqJu5eKEq7PQ5MgNayPkzZ9lEKddxDUEqVt4v7VHe
q9jh6B0CgYkA8F9wKifh8v2QqHm36bDSI1yTXA24DMdhd6ZYnuVPdg7/jSFv7yEn
7T3QgsGusm6b63e2qmtmr/otD4wF73ViXh4t7gDhgK4aK9+cqGbBl4c5+D+FomBw
2+OcPutYc+mc+tJhmz89Pt1izo1ANL+v0BARLRWNEv0VTenvvD8bR1xDqH6QZtKy
AwJ5AKQ9CWZc9o/rMSrRBcKxNwiPp+MSGTtqyYyBCW5Z5RYxAClh7XLKyUnoCjqt
3XBXjspxkyK+zL20HLZPHjwVkd+EdgTuL8S3Fh+kkoDhzZXRn9w/bHmr72Sd03Mr
Ivjx1gOp9ydxKmZ0Xoz4Nm09ZYyKlhZpOyF8/wKBiHCZhIV64VFejqEdQ5XpCscd
2rnIg2sZCwNtnR3x9WMsa3HFNBYkxftQdZiK+jcDsW6AScVTQms1Gl6qDS27IzVL
leBj8T7CT+g1e8E3QYCmC/XKa+NAoh2fZdXjkS/bQ3oLi0WaPipwspRnfqg3Ezi4
DhO8gLVgcNZqu67HMRQgmGEZyLMiB9kCeEmz4l3/Zd5b6yqNtooSQOIkpXCvFIen
el8FIRhWWwnEX5Ayk/4ppn72FHEUyQS7JicPJLo46WRQSXo+sxC/lUC7DsNTqDgc
+V6l2eDgdAPBmH2cMK/BSqLaeKN8Pit9S09FnNYkYKStoCie0r3fCY0yO/w+qPx0
PQKBiDIxFmoF0F36wZy/9jVHKpddtGfapDLlH8D8zMrCp7dLPKpF3BFVgSvXI2hC
PDGLFbGbyP1j+Ax9QyD15ItIzyQMhxnIg6EIS11CcogkYMQwGMacYyTFpeUKK6Qc
RQzrO8T67t2Een/Dc42MglBt1ztcYC+O2Rp0Bh1pSJJZSAtzXMS/pAJPmSI=
-----END RSA PRIVATE KEY-----`)

func rsaDecrypt(secretStr string) (string, error) {
	block, _ := pem.Decode(privateKey)
	if block == nil {
		log.Printf("fail, decrypt pem.Decode err\n")
		return "", fmt.Errorf("pem.Decode err")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		log.Printf("fail, decrypt x509.ParsePKCS1PrivateKey err, err=%v\n", err)
		return "", err
	}
	data, err := hex.DecodeString(secretStr)
	if err != nil {
		log.Printf("fail, decrypt hex.DecodeString err, err=%v\n", err)
		return "", err
	}
	str, err := rsa.DecryptPKCS1v15(rand.Reader, priv, data)
	if err != nil {
		log.Printf("fail, decrypt rsa.DecryptPKCS1v15 err, err=%v\n", err)
		return "", err
	}
	return string(str), nil
}
