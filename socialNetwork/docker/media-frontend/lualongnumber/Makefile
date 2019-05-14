LUAJIT_INC = /usr/local/openresty/luajit/include/luajit-2.1
LUAJIT_LIB = /usr/local/openresty/luajit/lib

all:
	gcc -o liblualongnumber.so       -shared -fPIC -g lualongnumber.c longnumberutils.c -I$(LUAJIT_INC)

install:
	install liblualongnumber.so /usr/local/openresty/lualib

clean:
	rm -rf liblualongnumber.so