docker rm -f $(docker ps -q --filter ancestor=gougoule-ai-frontend)
docker build -t gougoule-ai-frontend .
docker run -d --network=host gougoule-ai-frontend:latest