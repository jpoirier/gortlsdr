#


all:
	gcc -Wall -c librtlsdr_moc.c -o librtlsdr_moc.o
	ar rcs librtlsdr_moc.a librtlsdr_moc.o
	CC="gcc -DMOCK_TEST" go build --ldflags '-extldflags "-static"' -o gortlsdr_moc.a ../rtlsdr.go ../exports.go
	go build main_moc.go

clean:
	rm -f *.o *.a



