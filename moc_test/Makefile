#


all:
	cd gortlsdr_moc && gcc -Wall -c librtlsdr_moc.c -o librtlsdr_moc.o
	cd gortlsdr_moc && ar rcs librtlsdr_moc.a librtlsdr_moc.o
	cd gortlsdr_moc && cp ../../rtlsdr.go .
	cd gortlsdr_moc && cp ../../exports.go .
	cd gortlsdr_moc && CC="gcc -DMOCK_TEST" go build -tags 'MOC_TEST' --ldflags '-extldflags "-static"' -o gortlsdr_moc.a rtlsdr.go exports.go
	go build main_moc.go
	cd gortlsdr_moc &&  rm ./rtlsdr.go ./exports.go

clean:
	rm -f main_moc ./gortlsdr_moc/*.o ./gortlsdr_moc/*.a ./gortlsdr_moc/rtlsdr.go ./gortlsdr_moc/exports.go



