# define env var
DATABASE_URL?=file:./tmp/dev.db
PRSIMA=go run github.com/steebchen/prisma-client-go

db-push:
	export DATABASE_URL=$(DATABASE_URL) && $(PRSIMA) db push

migration-create:
	export DATABASE_URL=$(DATABASE_URL) && $(PRSIMA) migrate dev

migration-deploy:
	export DATABASE_URL=$(DATABASE_URL) && $(PRSIMA) migrate deploy
