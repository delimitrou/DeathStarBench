LUAJIT_INC = /usr/local/openresty/luajit/include/luajit-2.1
LUAJIT_LIB = /usr/local/openresty/luajit/lib

all:
	gcc -o libluabitwise.so          -shared -fPIC -g luabitwise.c longnumberutils.c    -I$(LUAJIT_INC)
	gcc -o libluabpack.so  -D_LUA51_ -shared -fPIC -g luabpack.c longnumberutils.c      -I$(LUAJIT_INC)
	gcc -o liblualongnumber.so       -shared -fPIC -g lualongnumber.c longnumberutils.c -I$(LUAJIT_INC)
	# gcc -o libluasocket.so           -shared -fPIC -g luasocket.c usocket.c             -I$(LUAJIT_INC)

install:
	install libluabitwise.so    /usr/local/openresty/lualib
	install libluabpack.so      /usr/local/openresty/lualib
	install liblualongnumber.so /usr/local/openresty/lualib
	# install libluasocket.so     /usr/local/openresty/lualib

clean:
	rm -rf liblualongnumber.so libluabpack.so libluabitwise.so