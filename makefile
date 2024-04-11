.PHONY: build clean help

all: build

build:
	go build -v -o KittenCore main.go

clean:
	rm KittenCore
	go clean -i .

help:
	@echo "make：编译包及其依赖"
	@echo "make clean：移除编译及缓存文件"