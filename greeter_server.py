# Copyright 2015 gRPC authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
"""The Python implementation of the GRPC helloworld.Greeter server."""

from concurrent import futures
import logging
import time
import grpc
import helloworld_pb2
import helloworld_pb2_grpc
from pymongo import MongoClient

class Greeter(helloworld_pb2_grpc.GreeterServicer):
    def __init__(self, db):
        self.collection = db['DNS']

    def SayHello(self, request, context):
        # Exemple de document à insérer
        document = {
            "message": "Hello, %s!" % request.name,
            "timestamp": time.time() 
        }

        # Insertion du document dans la collection
        result = self.collection.insert_one(document)

        return helloworld_pb2.HelloReply(message="Hello, %s!" % request.name)


def serve():
    port = "50051"

    # Créer une connexion MongoDB une seule fois
    client = MongoClient('mongodb://localhost:27017/')
    db = client['CDS']

    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    helloworld_pb2_grpc.add_GreeterServicer_to_server(Greeter(db), server)

    server.add_insecure_port("[::]:" + port)
    server.start()
    print("Server started, listening on " + port)
    server.wait_for_termination()


if __name__ == "__main__":
    logging.basicConfig()
    serve()
