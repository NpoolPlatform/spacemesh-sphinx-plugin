pipeline {
  agent any
  environment {
    GOPROXY = 'https://goproxy.cn,direct'
  }
  tools {
    go 'go'
  }
  stages {
    stage('Clone') {
      steps {
        git(url: scm.userRemoteConfigs[0].url, branch: '$BRANCH_NAME', changelog: true, credentialsId: 'KK-github-key', poll: true)
      }
    }

    stage('Prepare') {
      steps {
        sh 'make deps'
      }
    }

    stage('Linting') {
      when {
        expression { BUILD_TARGET == 'true' }
      }
      steps {
        sh 'make verify'
      }
    }

    stage('Switch to current cluster') {
      when {
        anyOf{
          expression { BUILD_TARGET == 'true' }
        }
      }
      steps {
        sh 'cd /etc/kubeasz; ./ezctl checkout $TARGET_ENV'
      }
    }

    stage('Unit Tests') {
      when {
        expression { BUILD_TARGET == 'true' }
      }
      steps {
        sh (returnStdout: false, script: '''
          devboxpod=`kubectl get pods -A | grep development-box | head -n1 | awk '{print $2}'`
          servicename="sphinx-plugin"

          kubectl exec --namespace kube-system $devboxpod -- make -C /tmp/$servicename after-test || true
          kubectl exec --namespace kube-system $devboxpod -- rm -rf /tmp/$servicename || true
          kubectl cp ./ kube-system/$devboxpod:/tmp/$servicename

          # kubectl exec --namespace kube-system $devboxpod -- make -C /tmp/$servicename deps before-test test after-test
          kubectl exec --namespace kube-system $devboxpod -- rm -rf /tmp/$servicename
        '''.stripIndent())
      }
    }

    stage('Generate docker image for development') {
      when {
        expression { BUILD_TARGET == 'true' }
      }
      steps {
        sh 'make verify-build'
        sh 'DEVELOPMENT=development DOCKER_REGISTRY=$DOCKER_REGISTRY make sphinx-plugin-image'
      }
    }

    stage('Tag patch') {
      when {
        expression { TAG_PATCH == 'true' }
      }
      steps {
        sh(returnStdout: true, script: '''
          set +e
          revlist=`git rev-list --tags --max-count=1`
          rc=$?
          set -e
          if [ 0 -eq $rc ]; then
            tag=`git describe --tags $revlist`
            major=`echo $tag | awk -F '.' '{ print $1 }'`
            minor=`echo $tag | awk -F '.' '{ print $2 }'`
            patch=`echo $tag | awk -F '.' '{ print $3 }'`
            case $TAG_FOR in
              testing)
                patch=$(( $patch + $patch % 2 + 1 ))
                ;;
              production)
                patch=$(( $patch + 1 ))
                git reset --hard
                git checkout $tag
                ;;
            esac
            tag=$major.$minor.$patch
          else
            tag=0.1.1
          fi
          git tag -a $tag -m "Bump version to $tag"
        '''.stripIndent())

        withCredentials([gitUsernamePassword(credentialsId: 'KK-github-key', gitToolName: 'git-tool')]) {
          sh 'git push --tag'
        }
      }
    }

    stage('Tag minor') {
      when {
        expression { TAG_MINOR == 'true' }
      }
      steps {
        sh(returnStdout: true, script: '''
          set +e
          revlist=`git rev-list --tags --max-count=1`
          rc=$?
          set -e
          if [ 0 -eq $rc ]; then
            tag=`git describe --tags $revlist`
            major=`echo $tag | awk -F '.' '{ print $1 }'`
            minor=`echo $tag | awk -F '.' '{ print $2 }'`
            patch=`echo $tag | awk -F '.' '{ print $3 }'`
            minor=$(( $minor + 1 ))
            patch=1
            tag=$major.$minor.$patch
          else
            tag=0.1.1
          fi
          git tag -a $tag -m "Bump version to $tag"
        '''.stripIndent())

        withCredentials([gitUsernamePassword(credentialsId: 'KK-github-key', gitToolName: 'git-tool')]) {
          sh 'git push --tag'
        }
      }
    }

    stage('Tag major') {
      when {
        expression { TAG_MAJOR == 'true' }
      }
      steps {
        sh(returnStdout: true, script: '''
          set +e
          revlist=`git rev-list --tags --max-count=1`
          rc=$?
          set -e
          if [ 0 -eq $rc ]; then
            tag=`git describe --tags $revlist`
            major=`echo $tag | awk -F '.' '{ print $1 }'`
            minor=`echo $tag | awk -F '.' '{ print $2 }'`
            patch=`echo $tag | awk -F '.' '{ print $3 }'`
            major=$(( $major + 1 ))
            minor=0
            patch=1
            tag=$major.$minor.$patch
          else
            tag=0.1.1
          fi
          git tag -a $tag -m "Bump version to $tag"
        '''.stripIndent())

        withCredentials([gitUsernamePassword(credentialsId: 'KK-github-key', gitToolName: 'git-tool')]) {
          sh 'git push --tag'
        }
      }
    }

    stage('Generate docker image for testing or production') {
      when {
        expression { BUILD_TARGET == 'true' }
      }
      steps {
        sh(returnStdout: true, script: '''
          revlist=`git rev-list --tags --max-count=1`
          tag=`git describe --tags $revlist`
          git reset --hard
          git checkout $tag
        '''.stripIndent())
        sh 'make verify-build'
        sh 'DEVELOPMENT=other DOCKER_REGISTRY=$DOCKER_REGISTRY make sphinx-plugin-image'
      }
    }

    stage('Release docker image for development') {
      when {
        expression { RELEASE_TARGET == 'true' }
      }
      steps {
        sh 'TAG=latest DOCKER_REGISTRY=$DOCKER_REGISTRY make sphinx-plugin-release'
        sh(returnStdout: true, script: '''
          images=`docker images | grep entropypool | grep sphinx-plugin | grep none | awk '{ print $3 }'`
          for image in $images; do
            docker rmi $image -f
          done
        '''.stripIndent())
      }
    }

    stage('Release docker image for testing') {
      when {
        expression { RELEASE_TARGET == 'true' }
      }
      steps {
        sh(returnStdout: false, script: '''
          revlist=`git rev-list --tags --max-count=1`
          tag=`git describe --tags $revlist`

          set +e
          docker images | grep sphinx-plugin | grep $tag
          rc=$?
          set -e
          if [ 0 -eq $rc ]; then
            TAG=$tag DOCKER_REGISTRY=$DOCKER_REGISTRY make sphinx-plugin-release
          fi
        '''.stripIndent())
      }
    }

    stage('Release docker image for production') {
      when {
        expression { RELEASE_TARGET == 'true' }
      }
      steps {
        sh(returnStdout: false, script: '''
          revlist=`git rev-list --tags --max-count=1`
          tag=`git describe --tags $revlist`

          major=`echo $tag | awk -F '.' '{ print $1 }'`
          minor=`echo $tag | awk -F '.' '{ print $2 }'`
          patch=`echo $tag | awk -F '.' '{ print $3 }'`

          patch=$(( $patch - $patch % 2 ))
          tag=$major.$minor.$patch

          set +e
          docker images | grep sphinx-plugin | grep $tag
          rc=$?
          set -e
          if [ 0 -eq $rc ]; then
            TAG=$tag DOCKER_REGISTRY=$DOCKER_REGISTRY make sphinx-plugin-release
          fi
        '''.stripIndent())
      }
    }
  }
  post('Report') {
    fixed {
      script {
        sh(script: 'bash $JENKINS_HOME/wechat-templates/send_wxmsg.sh fixed')
     }
      script {
        // env.ForEmailPlugin = env.WORKSPACE
        emailext attachmentsPattern: 'TestResults\\*.trx',
        body: '${FILE,path="$JENKINS_HOME/email-templates/success_email_tmp.html"}',
        mimeType: 'text/html',
        subject: currentBuild.currentResult + " : " + env.JOB_NAME,
        to: '$DEFAULT_RECIPIENTS'
      }
     }
    success {
      script {
        sh(script: 'bash $JENKINS_HOME/wechat-templates/send_wxmsg.sh successful')
     }
      script {
        // env.ForEmailPlugin = env.WORKSPACE
        emailext attachmentsPattern: 'TestResults\\*.trx',
        body: '${FILE,path="$JENKINS_HOME/email-templates/success_email_tmp.html"}',
        mimeType: 'text/html',
        subject: currentBuild.currentResult + " : " + env.JOB_NAME,
        to: '$DEFAULT_RECIPIENTS'
      }
     }
    failure {
      script {
        sh(script: 'bash $JENKINS_HOME/wechat-templates/send_wxmsg.sh failure')
     }
      script {
        // env.ForEmailPlugin = env.WORKSPACE
        emailext attachmentsPattern: 'TestResults\\*.trx',
        body: '${FILE,path="$JENKINS_HOME/email-templates/fail_email_tmp.html"}',
        mimeType: 'text/html',
        subject: currentBuild.currentResult + " : " + env.JOB_NAME,
        to: '$DEFAULT_RECIPIENTS'
      }
     }
    aborted {
      script {
        sh(script: 'bash $JENKINS_HOME/wechat-templates/send_wxmsg.sh aborted')
     }
      script {
        // env.ForEmailPlugin = env.WORKSPACE
        emailext attachmentsPattern: 'TestResults\\*.trx',
        body: '${FILE,path="$JENKINS_HOME/email-templates/fail_email_tmp.html"}',
        mimeType: 'text/html',
        subject: currentBuild.currentResult + " : " + env.JOB_NAME,
        to: '$DEFAULT_RECIPIENTS'
      }
     }
  }
}
