docker build --platform=linux/amd64 -f Dockerfile.devlake-official-style -t devlake-from-source-official:latest .
docker build --platform=linux/amd64 -f config-ui/Dockerfile -t devlake-config-ui-test:latest config-ui/


docker-compose -f docker-compose-test-new.yml up -d
docker-compose -f docker-compose-test-new.yml down
