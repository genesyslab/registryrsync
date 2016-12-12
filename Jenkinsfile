#!/usr/bin/env groovy

node {

  stage 'checkout'
  checkout scm

   stage "unit test"

   wrap([$class: 'AnsiColorBuildWrapper']) {
     withDockerContainer(image:'pitchanon/jenkins-golang') {
      sh 'go version'
      //  sh 'go test .'
      }
   }

   stage "build" {
     sh 'make builddocker'
   }


   stage 'push'
   imageName = 'genesyslab/registryrsync'
   docker.withRegistry(env.DOCKER_REG, env.DOCKER_REG_CRED_ID) {
     docker.image(imageName).push('latest')
   }
}
