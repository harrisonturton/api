openssl genrsa -out app.rsa 2048 # Private
openssl rsa -in app.rsa -pubout > app.rsa.pub # Public
