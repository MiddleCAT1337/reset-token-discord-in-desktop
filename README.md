# resettoken

โปรแกรมสำหรับ Windows ที่ช่วยสแกนหา Discord session token ที่เก็บอยู่ในเครื่องของเรา แล้วสั่ง logout ผ่าน API ของ Discord เพื่อให้ token นั้นใช้ต่อไม่ได้ รองรับทั้ง Discord แบบติดตั้งบนเครื่อง (Desktop, Canary, PTB) และการ login ผ่านเว็บในเบราว์เซอร์ (Chrome, Edge, Brave, Opera, Vivaldi)

ใช้ได้กับบัญชี Discord ของตัวเองบนเครื่องตัวเองเท่านั้น ไม่ได้ออกแบบมาให้เอาไปใช้กับเครื่องหรือบัญชีของคนอื่น

## การทำงานของโปรแกรม

1. โปรแกรมอ่านตัวเลือกจาก command line เช่น scan-only, all-sessions, yes
2. โมดูล paths รวบรวม path ที่จะสแกน ได้แก่โฟลเดอร์ Discord Desktop (Stable, Canary, PTB) และ Local Storage / Session Storage ของแต่ละโปรไฟล์เบราว์เซอร์
3. โมดูล scanner เปิดไฟล์ LevelDB ใน path เหล่านั้น แล้วใช้รูปแบบข้อความ (regex) หา string ที่น่าจะเป็น Discord user token จากนั้นกรองอีกครั้งด้วยการถอด base64 ส่วนแรกของ token ว่าเป็นเลข user id จริงหรือไม่
4. แต่ละ token ที่เจอ โมดูล discord เรียก GET /users/@me ของ Discord เพื่อดูว่ายังใช้ได้ไหม และดึงชื่อบัญชีมาแสดง
5. ถ้าไม่ใช่โหมด scan-only และผู้ใช้ยืนยัน (หรือใช้ -y) โปรแกรมจะเรียก POST /auth/logout เพื่อยกเลิก session ของ token นั้น หรือถ้าใช้ --all-sessions จะดึงรายการ session แล้วส่ง logout เป็นชุด
6. คำขอ HTTP ส่งจากเครื่องผู้ใช้ไปที่ discord.com โดยตรง ไม่มีการส่ง token ไปเซิร์ฟเวอร์อื่น
7. เมื่อจบงาน โปรแกรมรอให้กด Enter แล้วค่อยปิด

## ฟีเจอร์

- สแกนโฟลเดอร์ Local Storage และ Session Storage ของ Discord Desktop และโปรไฟล์เบราว์เซอร์
- ตรวจว่า token ที่เจอยังใช้งานได้จริงหรือไม่ โดยเรียก Discord API
- แสดงรายชื่อบัญชี แหล่งที่พบ token และ token แบบปิดบางส่วน (ไม่โชว์เต็ม)
- logout token ทีละรายการ หรือเลือก logout session ทั้งหมดของบัญชี (โหมด all-sessions)
- โหมด scan-only สำหรับดูอย่างเดียว ยังไม่ logout
- เมื่อทำงานเสร็จ โปรแกรมจะรอให้กด Enter ก่อนปิด

## เทคโนโลยีที่ใช้

- ภาษา Go (go 1.22 ขึ้นไป)
- เรียก HTTP API ของ Discord (ไม่มีไลบรารีภายนอกนอกจาก standard library)
- รองรับการรันบน Windows เป็นหลัก (path การสแกนออกแบบสำหรับ Windows)

## สิ่งที่ต้องมีก่อนติดตั้ง

- Windows 10 หรือใหม่กว่า
- Go เวอร์ชัน 1.22 ขึ้นไป ติดตั้งจาก https://go.dev/dl/
- การเชื่อมต่ออินเทอร์เน็ต ตอนตรวจ token และตอน logout

วิธีเช็กว่ามี Go แล้วหรือยัง เปิด PowerShell แล้วพิมพ์

```
go version
```

ถ้าเห็นเลขเวอร์ชัน เช่น go1.22 แปลว่าพร้อม build ได้

## วิธีติดตั้ง

1. เปิด PowerShell แล้วไปที่โฟลเดอร์โปรเจกต์ (ปรับ path ให้ตรงกับที่เก็บไฟล์จริง)

```
cd C:\Users\WhiteHorizon\Desktop\resettoken
```

2. ดาวน์โหลด dependency ของ Go (โปรเจกต์นี้ใช้แค่ standard library แต่คำสั่งนี้ช่วยให้แน่ใจว่า module พร้อม)

```
go mod tidy
```

3. build เป็นไฟล์ exe

```
go build -o resettoken.exe .
```

4. ถ้า build สำเร็จ จะมีไฟล์ resettoken.exe อยู่ในโฟลเดอร์เดียวกับ main.go

หมายเหตุ ถ้า build ไม่ผ่าน ให้เช็กว่า go version ตรงตามที่ระบุ และอยู่ในโฟลเดอร์ที่มีไฟล์ go.mod

## วิธีใช้งาน

### ขั้นที่ 1 เปิดโปรแกรมแบบดูอย่างเดียว (แนะนำครั้งแรก)

ยังไม่ logout ใช้แค่ดูว่าในเครื่องมี token อะไรที่ยังใช้ได้บ้าง

```
.\resettoken.exe --scan-only
```

สิ่งที่ควรเห็นบนหน้าจอ (ตัวอย่าง ตัวเลขอาจไม่ตรงกับเครื่องคุณ)

- บรรทัด Scanning ... locations บอกว่าสแกนกี่ตำแหน่ง
- บรรทัด Token-like strings ... Valid ... Expired ... สรุปจำนวน
- รายการบัญชี แหล่ง (Source) เช่น Chrome (Default) หรือ Discord Desktop
- Token แบบย่อ เช่น MTQ1OTk5...eV3E

ตอนจบจะขึ้น Press Enter to exit... กด Enter หนึ่งครั้งเพื่อปิด

### ขั้นที่ 2 logout token ทุกตัวที่ยังใช้ได้บนเครื่อง

แอป Discord และเบราว์เซอร์ที่ login อยู่จะหลุดออกหลัง logout สำเร็จ

```
.\resettoken.exe
```

โปรแกรมจะถามให้พิมพ์ yes แล้วกด Enter ถ้าไม่ต้องการให้ถาม ใช้

```
.\resettoken.exe -y
```

หรือ

```
.\resettoken.exe --yes
```

หลัง logout แต่ละบัญชีจะมีบรรทัด [OK] หรือ [FAIL] ตามด้วยชื่อผู้ใช้และข้อความสถานะ

### ขั้นที่ 3 โหมด logout session ทั้งหมดของแต่ละบัญชี (ตัวเลือก)

ใช้เมื่ออยากพยายามยกเลิก session หลายตัวของบัญชีเดียวกันผ่าน API มากกว่า logout แค่ token ที่เจอในไฟล์

```
.\resettoken.exe --all-sessions -y
```

ถ้าบัญชีเปิด MFA ไว้ อาจได้ข้อความว่าต้องยืนยัน MFA ในกรณีนั้นให้ logout ทีละ token แบบปกติ หรือเปลี่ยนรหัสผ่านใน Discord เพื่อยกเลิกทุกที่

### สรุปตัวเลือกคำสั่ง (flags)

- --scan-only ดูรายการ token ที่ใช้ได้ ไม่ logout
- -y หรือ --yes ข้ามคำถามยืนยันก่อน logout
- --all-sessions พยายาม logout session ทั้งหมดต่อบัญชี (อาจต้อง MFA)

### ถ้าสแกนแล้วไม่เจอ token

Discord เวอร์ชันใหม่อาจเข้ารหัส token ในเครื่อง ทำให้สแกนไม่เจอ วิธีที่แน่นอนกว่าคือ logout จากแอป Discord เอง หรือเข้า discord.com แล้วเปลี่ยนรหัสผ่าน ซึ่งจะยกเลิก session ทุกอุปกรณ์

### ข้อจำกัดที่ควรรู้

- โปรแกรมยกเลิกได้เฉพาะ token ที่หาเจอบนเครื่องนี้ ไม่ได้ logout ทุกอุปกรณ์ทั่วโลกให้อัตโนมัติ
- โปรแกรมส่งคำขอไปที่ discord.com จากเครื่องคุณเท่านั้น ไม่มีการส่ง token ไปเซิร์ฟเวอร์อื่น
- การเก็บหรือแชร์ Discord token ผิดกติกา Discord และเสี่ยงต่อการถูกแฮกบัญชี

## โครงสร้างโฟลเดอร์

```
resettoken
  main.go                 จุดเริ่มโปรแกรมและเมนูคำสั่ง
  go.mod                  ชื่อ module และเวอร์ชัน Go
  resettoken.exe          ไฟล์ที่ได้หลัง build (สร้างเอง)
  internal
    paths                 รายการ path ที่ใช้สแกน Desktop และเบราว์เซอร์
    scanner               อ่านไฟล์ LevelDB และดึง token ด้วย regex
    discord               เรียก Discord API ตรวจ user และ logout
    console               ตั้งค่า console บน Windows ให้แสดงผลได้
```
