EXT :=
ifeq ($(OS),Windows_NT)
	EXT = .exe
endif

vpn:
	@echo "install vpn for $(OS)"
	go install ./cmd/vpn

slist:
	@echo "building slist for linux"
	go build -o ./bin/slist ./cmd/slist

gitlab:
	@echo "install gitlab for $(OS)"
	go install ./cmd/gitlab

otp:
	@echo "install otp for $(OS)"
	go install ./cmd/otp
