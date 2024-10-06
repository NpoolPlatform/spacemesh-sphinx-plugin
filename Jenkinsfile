pipeline {
  agent any
  environment {
    GOPROXY='https://goproxy.cn,direct'
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
        sh 'rm -rf ./output/*'
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
          servicename="sphinx-plugin-p2"

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
        sh 'DOCKER_REGISTRY=$DOCKER_REGISTRY make sphinx-plugin-p2-image'
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
          if [ 0 -eq $rc -a x"$revlist" != x ]; then
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
          if [ 0 -eq $rc -a x"$revlist" != x ]; then
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
          if [ 0 -eq $rc -a x"$revlist" != x ]; then
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
          set +e
          revlist=`git rev-list --tags --max-count=1`
          rc=$?
          set -e
          if [ 0 -eq $rc -a x"$revlist" != x ]; then
            tag=`git describe --tags $revlist`
            git reset --hard
            git checkout $tag
          fi
        '''.stripIndent())
        sh 'make verify-build'
        sh 'DOCKER_REGISTRY=$DOCKER_REGISTRY make sphinx-plugin-p2-image'
      }
    }

    stage('Release docker image') {
      when {
        expression { RELEASE_TARGET == 'true' }
      }
      steps {
        sh(returnStdout: false, script: '''
          branch=latest
          if [ "x$BRANCH_NAME" != "xmaster" ]; then
            branch=`echo $BRANCH_NAME | awk -F '/' '{ print $2 }'`
          fi
          set +e
          docker images | grep sphinx-plugin-p2 | grep $branch
          rc=$?
          set -e
          if [ 0 -eq $rc ]; then
            DOCKER_REGISTRY=$DOCKER_REGISTRY make sphinx-plugin-p2-release
          fi
          images=`docker images | grep entropypool | grep sphinx-plugin-p2 | grep none | awk '{ print $3 }'`
          for image in $images; do
            docker rmi $image -f
          done
        '''.stripIndent())
      }
    }
    stage('Deploy for development') {
      when {
        expression { DEPLOY_TARGET == 'true' }
        expression { TARGET_ENV ==~ /.*development.*/ }
      }
      steps {
        sh(returnStdout: false, script: '''
          branch=latest
          if [ "x$BRANCH_NAME" != "xmaster" ]; then
            branch=`echo $BRANCH_NAME | awk -F '/' '{ print $2 }'`
          fi
          sed -i "s/sphinx-plugin-p2:latest/sphinx-plugin-p2:$branch/g" cmd/sphinx-plugin-p2/k8s/01-sphinx-plugin-p2.yaml
          sed -i "s/uhub.service.ucloud.cn/$DOCKER_REGISTRY/g" cmd/sphinx-plugin-p2/k8s/01-sphinx-plugin-p2.yaml
          export ENV_COIN_LOCAL_API=$ENV_COIN_LOCAL_API
          export ENV_COIN_PUBLIC_API=$ENV_COIN_PUBLIC_API
          export ENV_COIN_JSONRPC_LOCAL_API=$ENV_COIN_JSONRPC_LOCAL_API
          export ENV_COIN_JSONRPC_PUBLIC_API=$ENV_COIN_JSONRPC_PUBLIC_API
          export ENV_COIN_TYPE=$ENV_COIN_TYPE
          export ENV_COIN_NET=$ENV_COIN_NET
          export ENV_PROXY=$ENV_PROXY
          export ENV_SYNC_INTERVAL=$ENV_SYNC_INTERVAL
          export ENV_POSITION=$ENV_POSITION
          export ENV_WAN_IP=$ENV_WAN_IP
          export ENV_CONTRACT=$ENV_CONTRACT
          export ENV_CHAIN_ID=$ENV_CHAIN_ID
          export ENV_CHAIN_NICKNAME=$ENV_CHAIN_NICKNAME
          sed -i "s/sphinx-plugin-p2-coin/sphinx-plugin-p2-${ENV_COIN_TYPE}-${ENV_COIN_NET}/g" cmd/sphinx-plugin-p2/k8s/00-configmap.yaml
          sed -i "s/sphinx-plugin-p2-coin/sphinx-plugin-p2-${ENV_COIN_TYPE}-${ENV_COIN_NET}/g" cmd/sphinx-plugin-p2/k8s/01-sphinx-plugin-p2.yaml
          envsubst < cmd/sphinx-plugin-p2/k8s/00-configmap.yaml | kubectl apply -f -
          export PROXY_HOST_CONFIG=$PROXY_HOST_CONFIG
          envsubst < cmd/sphinx-plugin-p2/k8s/proxy-host-config.yaml | kubectl apply -f -
          make deploy-to-k8s-cluster
        '''.stripIndent())
      }
    }
    stage('Deploy for testing') {
      when {
        expression { DEPLOY_TARGET == 'true' }
        expression { TARGET_ENV ==~ /.*testing.*/ }
      }
      steps {
        sh(returnStdout: false, script: '''
          set +e
          revlist=`git rev-list --tags --max-count=1`
          rc=$?
          set -e
          if [ ! 0 -eq $rc -o x"$revlist" == x]; then
            exit 0
          fi
          tag=`git tag --sort=-v:refname | grep [1\\|3\\|5\\|7\\|9]$ | head -n1`

          git reset --hard
          git checkout $tag
          sed -i "s/sphinx-plugin-p2:latest/sphinx-plugin-p2:$tag/g" cmd/sphinx-plugin-p2/k8s/01-sphinx-plugin-p2.yaml
          sed -i "s/uhub.service.ucloud.cn/$DOCKER_REGISTRY/g" cmd/sphinx-plugin-p2/k8s/01-sphinx-plugin-p2.yaml
          sed -i "s/imagePullPolicy: Always/imagePullPolicy: IfNotPresent/g" cmd/sphinx-plugin-p2/k8s/01-sphinx-plugin-p2.yaml
          export ENV_COIN_LOCAL_API=$ENV_COIN_LOCAL_API
          export ENV_COIN_PUBLIC_API=$ENV_COIN_PUBLIC_API
          export ENV_COIN_JSONRPC_LOCAL_API=$ENV_COIN_JSONRPC_LOCAL_API
          export ENV_COIN_JSONRPC_PUBLIC_API=$ENV_COIN_JSONRPC_PUBLIC_API
          export ENV_COIN_TYPE=$ENV_COIN_TYPE
          export ENV_COIN_NET=$ENV_COIN_NET
          export ENV_PROXY=$ENV_PROXY
          export ENV_SYNC_INTERVAL=$ENV_SYNC_INTERVAL
          export ENV_POSITION=$ENV_POSITION
          export ENV_WAN_IP=$ENV_WAN_IP
          export ENV_CONTRACT=$ENV_CONTRACT
          export ENV_CHAIN_ID=$ENV_CHAIN_ID
          export ENV_CHAIN_NICKNAME=$ENV_CHAIN_NICKNAME
          sed -i "s/sphinx-plugin-p2-coin/sphinx-plugin-p2-${ENV_COIN_TYPE}-${ENV_COIN_NET}/g" cmd/sphinx-plugin-p2/k8s/00-configmap.yaml
          sed -i "s/sphinx-plugin-p2-coin/sphinx-plugin-p2-${ENV_COIN_TYPE}-${ENV_COIN_NET}/g" cmd/sphinx-plugin-p2/k8s/01-sphinx-plugin-p2.yaml
          envsubst < cmd/sphinx-plugin-p2/k8s/00-configmap.yaml | kubectl apply -f -
          export PROXY_HOST_CONFIG=$PROXY_HOST_CONFIG
          envsubst < cmd/sphinx-plugin-p2/k8s/proxy-host-config.yaml | kubectl apply -f -
          make deploy-to-k8s-cluster
        '''.stripIndent())
      }
    }
    stage('Deploy for production') {
      when {
        expression { DEPLOY_TARGET == 'true' }
        expression { TARGET_ENV ==~ /.*production.*/ }
      }
      steps {
        sh(returnStdout: false, script: '''
          set +e
          taglist=`git rev-list --tags`
          rc=$?
          set -e
          if [ ! 0 -eq $rc -o x"$revlist" == x]; then
            exit 0
          fi
          tag=`git tag --sort=-v:refname | grep [0\\|2\\|4\\|6\\|8]$ | head -n1`
          git reset --hard
          git checkout $tag
          sed -i "s/sphinx-plugin-p2:latest/sphinx-plugin-p2:$tag/g" cmd/sphinx-plugin-p2/k8s/01-sphinx-plugin-p2.yaml
          sed -i "s/uhub.service.ucloud.cn/$DOCKER_REGISTRY/g" cmd/sphinx-plugin-p2/k8s/01-sphinx-plugin-p2.yaml
          sed -i "s/imagePullPolicy: Always/imagePullPolicy: IfNotPresent/g" cmd/sphinx-plugin-p2/k8s/01-sphinx-plugin-p2.yaml
          export ENV_COIN_LOCAL_API=$ENV_COIN_LOCAL_API
          export ENV_COIN_PUBLIC_API=$ENV_COIN_PUBLIC_API
          export ENV_COIN_JSONRPC_LOCAL_API=$ENV_COIN_JSONRPC_LOCAL_API
          export ENV_COIN_JSONRPC_PUBLIC_API=$ENV_COIN_JSONRPC_PUBLIC_API
          export ENV_COIN_TYPE=$ENV_COIN_TYPE
          export ENV_COIN_NET=$ENV_COIN_NET
          export ENV_PROXY=$ENV_PROXY
          export ENV_SYNC_INTERVAL=$ENV_SYNC_INTERVAL
          export ENV_POSITION=$ENV_POSITION
          export ENV_WAN_IP=$ENV_WAN_IP
          export ENV_CONTRACT=$ENV_CONTRACT
          export ENV_CHAIN_ID=$ENV_CHAIN_ID
          export ENV_CHAIN_NICKNAME=$ENV_CHAIN_NICKNAME
          sed -i "s/sphinx-plugin-p2-coin/sphinx-plugin-p2-${ENV_COIN_TYPE}-${ENV_COIN_NET}/g" cmd/sphinx-plugin-p2/k8s/00-configmap.yaml
          sed -i "s/sphinx-plugin-p2-coin/sphinx-plugin-p2-${ENV_COIN_TYPE}-${ENV_COIN_NET}/g" cmd/sphinx-plugin-p2/k8s/01-sphinx-plugin-p2.yaml
          envsubst < cmd/sphinx-plugin-p2/k8s/00-configmap.yaml | kubectl apply -f -
          export PROXY_HOST_CONFIG=$PROXY_HOST_CONFIG
          envsubst < cmd/sphinx-plugin-p2/k8s/proxy-host-config.yaml | kubectl apply -f -
          make deploy-to-k8s-cluster
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
