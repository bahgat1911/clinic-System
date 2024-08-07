#!/bin/bash

container_exists() {
  local container_name="$1"
  docker ps -a --format '{{.Names}}' | grep -q "^$container_name$"
}

if container_exists mydatabase; then
  echo "mydatabase exists. Starting..."
  docker start mydatabase
else
  echo "mydatabase does not exist. Creating..."
docker run -d --name mydatabase --net mynetwork1 -p 3306:3306 database
fi

if container_exists mybackend; then
  echo "mybackend exists. Starting..."
  docker start mybackend
else
  echo "mybackend does not exist. Creating..."
  docker run -d --name mybackend --net mynetwork1 -p 8080:8080 backend
fi

if container_exists myfrontend; then
  echo "myfrontend exists. Starting..."
  docker start myfrontend
else
  echo "myfrontend does not exist. Creating..."
  docker build -t your-frontend-image .
docker run -d --name myfrontend --net mynetwork1 -p 3000:3000 frontend
fi

docker inspect mynetwork1
docker port mybackend
docker port myfrontend
