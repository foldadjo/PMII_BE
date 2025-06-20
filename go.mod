module handler

go 1.21

require (
	github.com/gofiber/fiber/v2 v2.52.2
	github.com/golang-jwt/jwt/v5 v5.2.1
	go.mongodb.org/mongo-driver v1.14.0
	golang.org/x/crypto v0.21.0
)

require (
	github.com/aws/aws-lambda-go v1.49.0 // indirect
	github.com/valyala/fasthttp v1.51.0 // indirect
)

require (
	github.com/andybalholm/brotli v1.0.5 // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/google/uuid v1.5.0 // indirect
	github.com/klauspost/compress v1.17.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/montanaflynn/stats v0.0.0-20171201202039-1bf9dbcd8cbe // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

require (
	github.com/foldadjo/PMII_BE v0.0.0-unpublished
	github.com/vercel/go-bridge v0.0.0-20221108222652-296f4c6bdb6d
)

replace github.com/foldadjo/PMII_BE => ./
