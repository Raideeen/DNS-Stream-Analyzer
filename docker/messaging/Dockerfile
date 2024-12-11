# 
# Custom Kafka image with administration binaries
#

# Build the Kafka application
FROM apache/kafka:latest AS broker-build 

# Set destination for COPY
WORKDIR /app 

# Add the Kafka administration commands files
ADD https://github.com/Kitware/CMake/releases/download/v3.27.6/cmake-3.27.6-linux-x86_64.sh /tmp/

# 