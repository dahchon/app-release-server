# define env var
DATABASE_URL?=file:./tmp/dev.db

db-push:
	export DATABASE_URL=$(DATABASE_URL) && go run github.com/steebchen/prisma-client-go db push
