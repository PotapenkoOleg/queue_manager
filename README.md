# queue_manager
Service to manage data availability without blocking Kafka queue

## Resolve kafka not found issue 

### Windows

[stackoverflow](https://stackoverflow.com/questions/76239071/does-confluent-kafka-go-package-compatible-with-ubuntu-22-04)


- [confluent-kafka-go](https://github.com/confluentinc/confluent-kafka-go) is a lightweight wrapper around [librdkafka](https://github.com/confluentinc/librdkafka) - a finely tuned C client
- It needs to be compiled with a GCC toolchain
- Windows does not include GCC by default, you must install a distribution like MinGW-w64
- Download MSYS2 installer from [MSYS2 website](https://www.msys2.org/#installation) and run it  
- Install GCC: Open the MSYS2 UCRT64 terminal and run: `pacman -S mingw-w64-x86_64-gcc`
- Set Environment Variables: Add `C:\msys64\mingw64\bin` to your system PATH, so GO can find the GCC executable
- On Windows 11 open "Settings" and type `env` in search to find environment variables
- Verify: Restart your terminal and run `gcc --version` and `go env CGO_ENABLED` (it should be 1).  
- Enable GCC in GO env `go env -w CGO_ENABLED="1"` if previous command returned 0
