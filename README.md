# grpc-poc

This is to try out a grpc model with managed endpoints

## Steps
  1. Clone the repo
  1. Create an AML workspace. Remember the ACR name that was created for the workspace.
  1. Build the GRPC model image
      ```sh
      docker build -t yourworkspaceacr.azurecr.io/joongrpc:1 -f ./Dockerfile .
      docker push yourworkspaceacr.azurecr.io/joongrpc:1
      ```
  1. Create the endpoint 
      ```sh
      az ml online-endpoint create --name $ENDPOINT_NAME -f ./deployment/http2Endpoint.yml
      ```
  1. Change the image to point to the workspace ACR image tag. Create the deployment and also update the traffic values
      ```sh
      az ml online-deployment create --name blue --endpoint $ENDPOINT_NAME -f ./deployment/http2Deployment.yml -g $RESOURCE_GROUP -w $WORKSPACE_NAME --all-traffic
      ```
      If you are using a different ACR you will have to make the change so that your deployment has access to pull the image from your ACR.

