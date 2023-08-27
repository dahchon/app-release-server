-- CreateTable
CREATE TABLE "app_releases" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "appName" TEXT NOT NULL,
    "appVersion" TEXT NOT NULL,
    "appBuild" TEXT NOT NULL,
    "gitCommit" TEXT NOT NULL,
    "mainFileName" TEXT NOT NULL,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL,
    "title" TEXT,
    "published" BOOLEAN NOT NULL DEFAULT true,
    "description" TEXT,
    "downloadCount" INTEGER NOT NULL DEFAULT 0,
    "uploaderIP" TEXT
);

-- CreateIndex
CREATE UNIQUE INDEX "app_releases_appName_appVersion_appBuild_key" ON "app_releases"("appName", "appVersion", "appBuild");
