<start backend with nodemon>
nodemon --exec "go run main.go" --ext go

<go to db in docker>
docker exec -it <container_name_or_id> psql -U <username> -d <database>

<stop all running Docker containers>
docker stop $(docker ps -q)

<delete all data docker volume>
docker-compose down -v

<List all volumes (optional preview)>
docker volume ls

<Delete all volumes>
docker volume prune -f

<Clean Everything (Containers + Volumes + Networks + Images)>
docker system prune -a --volumes -f


<create project>
go mod init warehouse-store
go mod tidy (update dependencies)
