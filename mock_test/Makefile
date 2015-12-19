#

all:
	# cd gortlsdr && gcc -fPIC -shared -Wl,-soname,librtlsdr.so.0 -lpthread -o librtlsdr.so.0 librtlsdr.c
	cd gortlsdr && gcc -Wall -c librtlsdr.c -lpthread -o librtlsdr.o
	cd gortlsdr && ar rcs librtlsdr.a librtlsdr.o
	cd gortlsdr && cp ../../rtlsdr.go .
	cd gortlsdr && cp ../../exports.go .
	cd gortlsdr && CC="gcc -Dmock" go build -o gortlsdr.a rtlsdr.go exports.go
	go build --ldflags '-extldflags "-L./gortlsdr"' main.go
	rm ./gortlsdr/rtlsdr.go ./gortlsdr/exports.go
	./main
clean:
	rm -f main ./gortlsdr/*.o ./gortlsdr/*.a ./gortlsdr/rtlsdr.go ./gortlsdr/exports.go ./gortlsdr/*.0 ./gortlsdr/*.so



