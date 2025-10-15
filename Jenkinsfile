pipeline {
    agent any

    environment {
        APP_NAME = "kairyu"
        DOCKER_IMAGE = "starwalkn/kairyu"
    }

    stages {

        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Docker Build & Push') {
            steps {
                withCredentials([usernamePassword(credentialsId: 'dockerhub-creds', usernameVariable: 'DOCKER_USER', passwordVariable: 'DOCKER_PASS')]) {
                    sh '''
                        export DOCKER_CONFIG=$WORKSPACE/.docker
                        mkdir -p $DOCKER_CONFIG

                        echo "$DOCKER_PASS" | /usr/local/bin/docker login -u "$DOCKER_USER" --password-stdin




                        /usr/local/bin/docker buildx build --platform linux/amd64,linux/arm64 -f ./build/Dockerfile \
                          -t $DOCKER_IMAGE:latest --push .


                        rm -rf $DOCKER_CONFIG
                    '''
                }
            }
        }

    }

    post {
        success {
            echo "✅ Docker image built and pushed successfully!"
        }
        failure {
            echo "❌ Build or push failed!"
        }
    }
}
