curl -X GET "localhost:8080/testOnlyBindQuery?name=yunfen&address=xyz"

{"Name":"yunfen","address":"xyz"}

类型	测试命令
JSON	curl -X POST 'http://localhost:8080/loginJSON' -v -d '{"user":"manu", "password":"123"}'
XML	curl -X POST "http://localhost:8080/loginXML" -v -d '<?xml version="1.0" encoding="UTF-8"?><root><user>manu</user><password>123</password></root>'
form	curl -X POST "http://localhost:8080/loginForm" -v -d 'user=manu&password=123'

curl -X POST 'http://localhost:8080/login' -v -d '{"email":"ccitt@qq.com", "password":"my123"}'

curl -X POST 'http://localhost:8080/signup' -v -d '{"name":"ccitt", "email":"ccitt@qq.com", "password":"my123", "phone_number":"1331234"}'
