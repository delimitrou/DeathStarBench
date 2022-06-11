@REM SBT launcher script
call .\sbt.bat microservice_1/docker:publishLocal
call .\sbt.bat microservice_2/docker:publishLocal
docker-compose up