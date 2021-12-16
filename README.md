# camunda-cloud-operator
Christmas Hack Day Project (2021): Build an Kubernetes Operator to deploy Camunda Cloud services

## Idea

Building an open-source Camunda Cloud operator which can be used and extended by the community to deploy and run Camunda Cloud services, like Zeebe, Operate, Tasklist, Optimize etc. For each service we build/add an extra controller and CRD which is bundled in the operator.

## Goals

### Goals for the Christmas Hackdays 2021

The goals for the Christmas Hackdays are simple: get something "useful" running. Ideally at the end of the day we are able to deploy a Zeebe Cluster with a new CRD.

Possible CRD (example):

```yaml
zeebe:
  cluster:
    size: 3
    replicationFactor: 3
    partition: 3
    envs:
      - key: value
      
    resources:
      memory: 
        limit:
        request:
      cpu:
        limit:
        request:  
  gateway:
    embedded: false  
    envs:
      - key: value
    resources: #optional
       memory: 
        limit:
        request:
      cpu:
        limit:
        request:
```

### Future

In the future we extend the operator with different controllers/crd's for each service.

## Resources

Useful resources to check out:

 * https://book.kubebuilder.io/introduction.html
 * https://cloud.redhat.com/blog/kubernetes-operators-best-practices
