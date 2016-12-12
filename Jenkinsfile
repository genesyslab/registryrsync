#!/usr/bin/env groovy

node {

  stage 'checkout'
  checkout scm

  wrap
   checkout scm: [$class: 'MercurialSCM', source: 'https://hg.genesyslab.com/hgweb.cgi/gir_rp_nodejs',  credentialsId: 'hgservicecreds'], poll: false
   stage "unit test"

   wrap([$class: 'AnsiColorBuildWrapper']) {
     withDockerContainer(image:'pitchanon/jenkins-golang') {
       sh 'go test .'
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
