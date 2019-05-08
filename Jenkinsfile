pipeline {
    agent {
        docker {
            image 'teambitflow/golang-docker:1.11-stretch'
            args '-v /root/.goroot:/go -v /root/.docker:/root/.docker'
        }
    }
    environment {
        docker_image = 'teambitflow/go-bitflow'
    }
    stages {
        stage('Build & test') { 
            steps {
                sh '''
                    go install ./...
                    go test -v ./...
                '''
            }
            /*
            post {
                always {
                    // TODO: capture test results. Enable coverage and capture report.
                    // TODO: add static code analysis stage
                }
            }
            */
        }
        stage('Build container') {
            when {
                branch 'master'
            }
            steps {
                script {
                    sh 'docker build -t $docker_image:build-$BUILD_NUMER -t $docker_image:latest .'
                }
            }
            post {
                success {
                    sh '''
                        docker push $docker_image:build-$BUILD_NUMBER
                        docker push $docker_image:latest'
                    '''
               }
            }
        }
    }
}

