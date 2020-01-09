
protoc --lua_out=.\Pbc --plugin=protoc-gen-lua=.\protoc-gen-lua\protoc-gen-lua_enum.bat .\GameRoom.proto
protoc --descriptor_set_out=.\Pbc\GameRoomMsg.pb.bytes .\GameRoomMsg.proto
Pause

  

 