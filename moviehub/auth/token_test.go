package auth

import (
    "encoding/base64"
    "strings"
    "testing"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

var testSecret = []byte("กุญแจลับสำหรับทดสอบเท่านั้น-ยาวพอแน่นอน")

func sign(t *testing.T, method jwt.SigningMethod, key any, claims jwt.MapClaims) string {
    t.Helper()
    signed, err := jwt.NewWithClaims(method, claims).SignedString(key)
    if err != nil {
        t.Fatal(err)
    }
    return signed
}

func TestForgedTickets(t *testing.T) {
    issuer := &Issuer{Secret: testSecret, TTL: time.Hour}
    genuine, err := issuer.Issue(1, "member")
    if err != nil {
        t.Fatal(err)
    }
    now := time.Now()
    good := jwt.MapClaims{"sub": "1", "role": "admin", "exp": now.Add(time.Hour).Unix()}

    // เปลี่ยนสิทธิ์ในเนื้อความโดยไม่เซ็นใหม่
    parts := strings.Split(genuine, ".")
    payload, err := base64.RawURLEncoding.DecodeString(parts[1])
    if err != nil {
        t.Fatal(err)
    }
    payload = []byte(strings.Replace(string(payload), "member", "admin", 1))
    edited := parts[0] + "." + base64.RawURLEncoding.EncodeToString(payload) + "." + parts[2]

    cases := []struct {
        name string
        token string
    }{
        {"แก้เนื้อความ", edited},
        {"เซ็นด้วยกุญแจอื่น", sign(t, jwt.SigningMethodHS256, []byte("กุญแจของคนอื่น-ยาวพอเหมือนกันนะ"), good)},
        {"อ้างว่าไม่ต้องเซ็น", sign(t, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, good)},
        {"กุญแจถูกแต่วิธีเซ็นอื่น", sign(t, jwt.SigningMethodHS512, testSecret, good)},
        {"ไม่มีวันหมดอายุ", sign(t, jwt.SigningMethodHS256, testSecret, jwt.MapClaims{"sub":"1", "role":"admin"})},
        {"หมดอายุแล้ว", sign(t, jwt.SigningMethodHS256, testSecret, jwt.MapClaims{"sub":"1", "role":"member", "exp":now.Add(-time.Minute).Unix()})},
    }
    if id, role, err := issuer.Parse(genuine); err != nil || id != 1 || role != "member" {
        t.Fatalf("ตั๋วจริงต้องผ่าน: id=%d role=%q err=%v", id, role, err)
    }
    t.Log("ตั๋วจริง ผ่าน")
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            _, _, err := issuer.Parse(tc.token)
            if err == nil {
                t.Fatal("ตั๋วปลอมผ่านเข้ามาได้")
            }
            t.Logf("ถูกปฏิเสธ: %v", err)
        })
    }
}
