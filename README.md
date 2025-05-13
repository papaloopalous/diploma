# diploma

to do:

сделать graceful stop для healthcheck

добавить отправку ключа с помощью диффи хелмана

сделать вход по алгоритму похожему на керберос

убрать jwt и поменять его на обычное шифрование

убрать дублирование ошибок в логгере в хендлерах, например({"level":"error","ts":1747046793.870688,"caller":"cmd/main.go:39","msg":"rpc error: code = Unimplemented desc = method GetUsersByIDs not implemented","service":"user","userID":"a5adb249-87ee-496f-925c-b742a90635fd","details":"rpc error: code = Unimplemented desc = method GetUsersByIDs not implemented"})