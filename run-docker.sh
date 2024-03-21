#!/bin/bash

docker run \
-p 8000:8000 \
-v $(pwd)/db:/app/db \
-it adai
