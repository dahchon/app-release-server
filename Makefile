# define env var
export DATABASE_URL=file:./tmp/dev.db

db-push:
	go run github.com/steebchen/prisma-client-go db push
