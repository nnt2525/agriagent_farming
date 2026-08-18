package main

import (
    "log"

    "github.com/gofiber/fiber/v2"
)

func main() {
    // สร้างแอปพลิเคชัน Fiber
    app := fiber.New()

    // สร้าง Route พื้นฐาน
    app.Get("/", func(c *fiber.Ctx) error {
        return c.SendString("Hello, World! From Fiber 🚀")
    })

    // รันเซิร์ฟเวอร์ที่พอร์ต 8080
    log.Fatal(app.Listen(":8080"))
}