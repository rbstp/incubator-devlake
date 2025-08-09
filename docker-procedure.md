docker build --platform=linux/amd64 -f Dockerfile.devlake-official-style -t devlake-from-source-official:latest .
docker build --platform=linux/amd64 -f config-ui/Dockerfile -t devlake-config-ui-with-testmo:latest config-ui/


docker-compose -f docker-compose-testmo-new.yml up -d
docker-compose -f docker-compose-testmo-new.yml down