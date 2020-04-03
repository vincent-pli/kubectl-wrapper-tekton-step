# kubectl-wrapper
Wrapper `kubectl` in container. it should work as `step` of [Tekton-pipeline](https://github.com/tektoncd/pipeline), the goal is to operate(create, delete, apply, patch, replace) on `k8s` resources.  

## Take a try
1. Build image  

    `make image TAG=V0.0.1`  

2. Push image to your favourite repo registry  

3. Run `yaml`s in `./deploy` in a `tekton` ready environment, don't forget to replace the image in `kubectl-deploy.yaml`  

