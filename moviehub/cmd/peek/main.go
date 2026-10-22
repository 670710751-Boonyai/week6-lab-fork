package main

import (
    "encoding/base64"
    "fmt"
    "os"
    "strings"
)

// peek เปิดอ่านส่วนหัวและเนื้อความของ JWT โดยไม่ต้องมีกุญแจลับ
// การอ่านได้ไม่ได้หมายความว่าลายเซ็นผ่านการตรวจสอบ
func main() {
    if len(os.Args) != 2 {
        fmt.Fprintln(os.Stderr, "วิธีใช้: go run ./cmd/peek $TOKEN")
        os.Exit(1)
    }
    parts := strings.Split(os.Args[1], ".")
    if len(parts) != 3 {
        fmt.Fprintln(os.Stderr, "JWT ต้องมี 3 ท่อน")
        os.Exit(1)
    }
    fmt.Println("ตั๋วมี", len(parts), "ท่อน")
    for i, name := range []string{"ส่วนหัว", "เนื้อความ"} {
        raw, err := base64.RawURLEncoding.DecodeString(parts[i])
        if err != nil {
            fmt.Println(name, "อ่านไม่ออก", err)
            continue
        }
        fmt.Printf("%s %s\n", name, raw)
    }
    fmt.Println("ลายเซ็นยาว", len(parts[2]), "ตัว (ไม่ได้ตรวจสอบลายเซ็น)")
}
