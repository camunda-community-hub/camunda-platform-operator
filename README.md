[![](https://img.shields.io/badge/Community%20Extension-An%20open%20source%20community%20maintained%20project-FF4700)](https://github.com/camunda-community-hub/community)
![Compatible with: Camunda Platform 8](https://img.shields.io/badge/Compatible%20with-Camunda%20Platform%208-0072Ce)

# Camunda Platform 8 Operator
Christmas Hack Day Project (2021): Build an Kubernetes Operator to deploy Camunda Platform 8 services

## Motiviation / Idea

We currently have open-source helm charts in order to deploy Camunda Platform 8 services, https://github.com/camunda/camunda-platform-helm. But to support further use cases, like deploying and maintaining lot of Zeebe clusters and doing this in a "simple" way (via CRD's) we need an Kubernetes Operator.

Of course part of the motiviation is to get the hands dirty and build up more knowledge about kubernetes and kubernetes operators.

So the idea is to build an open-source Camunda Platform 8 operator, which can be used and extended by the community to deploy and run Camunda Platform 8 services, like Zeebe, Operate, Tasklist, Optimize etc. For each service we want to build/add an extra controller and CRD which is bundled in the operator.



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
 * https://developers.redhat.com/blog/2019/10/07/write-a-simple-kubernetes-operator-in-java-using-the-fabric8-kubernetes-client#
