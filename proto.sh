#! 编译.proto文件为go、js、ts语言用的文件
cd msg
protoc --go_out=. *.proto
pbjs -t static-module -w commonjs -o CDXproto_msg.js *.proto
pbts -o CDXproto_msg.d.ts CDXproto_msg.js
echo succeed compiled