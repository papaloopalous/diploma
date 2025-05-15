# diploma

to do:

сделать graceful stop для healthcheck

сделать инит параметров через вайпер

пересмотреть все выводы ошибок, расставить пустые строки для разделения лог частей кода

сделать единые выводы для апи бд

проверить все функции

убрать дублирование ошибок в логгере в хендлерах, например({"level":"error","ts":1747046793.870688,"caller":"cmd/main.go:39","msg":"rpc error: code = Unimplemented desc = method GetUsersByIDs not implemented","service":"user","userID":"a5adb249-87ee-496f-925c-b742a90635fd","details":"rpc error: code = Unimplemented desc = method GetUsersByIDs not implemented"})