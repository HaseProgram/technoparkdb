# Databases course project for Technopark @mail.ru 

## Task
There is API without implementation: https://tech-db-forum.bozaro.ru/. We need to create application (any language) which will process requests with correct responses. After that we need to upgrade our application so it can process as many requests as it possible.
## Grades
Here are grades for our course (01.09.2017 - 20.01.2018). 1 point per RPS (Request per Second)

|DATE|... - 12.01.2018 | 12.01.2018 - 14.01.2018 | 15.01.2018 - ... |
|--|:--:|:--:|:--:|
| **RPS** | 50 | 40* |80

*There was server update which decresed it's perfomance.
## Result
Database fills in 5 minutes.
Perfomance test: 2669 RPS

## Launch (Docker)
This application is provided by Docker container. So you just need build and run:

    docker build -t d.zaytsev https://github.com/HaseProgram/technoparkdb.git
    docker run -p 5000:5000 --name d.zaytsev -t d.zaytsev

## Launch test
You can found test app here: [Application](https://github.com/bozaro/tech-db-forum#%D0%A4%D1%83%D0%BD%D0%BA%D1%86%D0%B8%D0%BE%D0%BD%D0%B0%D0%BB%D1%8C%D0%BD%D0%BE%D0%B5-%D1%82%D0%B5%D1%81%D1%82%D0%B8%D1%80%D0%BE%D0%B2%D0%B0%D0%BD%D0%B8%D0%B5)

Launching functional test:

    ./tech-db-forum func
Launching fill database:

    ./tech-db-forum fill
Launching perfomancetest:

    ./tech-db-forum perf
