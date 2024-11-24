# Miporin - みぽりん - the tank commander of ikukantai fleet

[![release](https://img.shields.io/badge/miporin--v1.0-log?style=flat&label=release&color=hotpink)]()
[![LICENSE](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](https://www.apache.org/licenses/LICENSE-2.0)
[![CloudNet2024](https://img.shields.io/badge/IEEE--CloudNet--2024-log?style=flat&label=publication&color=dodgerblue)](https://cloudnet2024.ieee-cloudnet.org)

[![Kubernetes](https://img.shields.io/badge/kubernetes-%23326ce5.svg?style=for-the-badge&logo=kubernetes&logoColor=white&link=https%3A%2F%2Fkubernetes.io)](https://kubernetes.io/)
[![Linux](https://img.shields.io/badge/Linux-FCC624?style=for-the-badge&logo=linux&logoColor=black)]()
[![Knative](https://img.shields.io/badge/knative-log?style=for-the-badge&logo=knative&logoColor=white&labelColor=%230865AD&color=%230865AD)](https://knative.dev/docs/)
[![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![Prometheus](https://img.shields.io/badge/Prometheus-E6522C?style=for-the-badge&logo=Prometheus&logoColor=white)](https://prometheus.io/)

`miporin`-chan is the extra-controller of `ikukantai`, working alongside and independently of Knative's controller.

![](docs/images/miporin_wp.jpg)

## 1. Motivation

To achieve the [goals](https://github.com/bonavadeur/ikukantai?tab=readme-ov-file#1-motivation) posed by the `ikukantai` Fleet, in addition to modifying Knative's source code, we needed a component acts as a controller that exploits the refined code inside Knative. In theory, we can develop additional logic in Knative's controller component. However, that will be more difficult than developing an extra external component for PoC purposes in the Laboratory (yaa, we work in the Laboratory, not Industry).

The name `miporin` is inspired by the character `Nishizumi Miho` from the anime `Girls und Panzer`. Miho is the tank commander, implying `miporin`'s leadership role in the `ikukantai` fleet (remember that Ooarai High School is located in an aircraft carrier, and, `ikukantai` is implied to be that ship). `miporin` is nickname given to Miho by her friends.

## 2. System Design

![](docs/images/design.png)

## 3. Installation

### 3.1. Requirement

+ [ikukantai](https://github.com/bonavadeur/ikukantai?tab=readme-ov-file#3-installation) Fleet is deployed, version >= 2.0
+ [ko build](https://ko.build/install/) is installed, version 0.16.0
+ [Go](https://go.dev/doc/install) is installed, version >= 1.22.4
+ [Docker]() is installed. `docker` command can be invoked without sudo

### 3.2. Installation

`miporin` is deployed in namespace **knative-serving**

```bash
kubectl apply -f config/miporin.yaml
```

### 3.3. Development

Firstly, modify image used by **deployment/miporin** in namespace **knative-serving** by image named `docker.io/bonavadeur/miporin:dev`. A new Pod miporin will be raised up due to the previous changes, and this Pod will be failed. Next, build your own image for development environment:

```bash
$ kubectl -n knative-serving patch deploy miporin --patch '{"spec":{"template":{"spec":{"containers":[{"name":"miporin","image":"docker.io/bonavadeur/miporin:dev"}]}}}}'
$ chmod +x ./build.sh
$ ./build.sh ful
```

Change Endpoint IP address to IP of your machine for running `miporin` by binary:

```yaml
# file ./config/localdev.yaml
apiVersion: discovery.k8s.io/v1
kind: EndpointSlice
metadata:
  name: miporin-localdev
  namespace: knative-serving
...
endpoints:
  - addresses:
    - "192.168.122.100" # change this to be your IP, example: 192.168.189.22
```

Some util commands

```bash
# grant execute permission to build.sh file
chmod +x ./build.sh
# run code directly by binary
./build.sh local
# run miporin as a container
./build.sh ful
# push miporin image to docker registry
./build.sh push <tag>
```

## 4. Author

Đào Hiệp - Bonavadeur - ボナちゃん  
The Future Internet Laboratory, Room E711 C7 Building, Hanoi University of Science and Technology, Vietnam.  
未来のインターネット研究室, C7 の E ７１１、ハノイ百科大学、ベトナム。  

![](docs/images/github-wp.png)  
