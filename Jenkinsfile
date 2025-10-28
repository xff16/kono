pipeline {
    agent any

    environment {
        APP_NAME = "branka"
        DOCKER_IMAGE = "starwalkn/branka"
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

                        /usr/local/bin/docker buildx version || (echo "🚨 Buildx not found!" && exit 1)

                        /usr/local/bin/docker buildx create --use --name multiarch-builder || true
                        /usr/local/bin/docker buildx inspect --bootstrap

                        /usr/local/bin/docker buildx build \
                            --platform linux/amd64,linux/arm64 \
                            -f ./build/Dockerfile \
                            -t starwalkn/bravka:latest \
                            --push .
                        /usr/local/bin/docker push $DOCKER_IMAGE:latest

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
