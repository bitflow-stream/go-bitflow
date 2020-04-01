pipeline {
    options {
        timeout(time: 1, unit: 'HOURS')
    }
    agent none
    environment {
        registry = 'bitflowstream/bitflow-pipeline'
        registryCredential = 'dockerhub'
    }

    stages {
        stage('Git') {
            agent {
                label 'master'
            }
            steps {
                script {
                    env.GIT_COMMITTER_EMAIL = sh(script: "git --no-pager show -s --format='%ae'", returnStdout: true).trim()
                }
            }
        }

        stage('Unit tests') {
            agent {
                docker {
                    image 'bitflowstream/golang-build:debian'
                    args '-v /tmp/go-mod-cache/debian:/go'
                }
            }
            steps {
                sh 'go clean -i -v ./...'
                sh 'go install -v ./...'
                sh 'rm -rf reports && mkdir -p reports'
                sh 'stdbuf -o 0 go test -v ./... -coverprofile=reports/test-coverage.txt | tee reports/test-output.txt' // Use stdbuf to disable buffering of the tee output
                sh 'cat reports/test-output.txt | go-junit-report -set-exit-code > reports/test.xml'
                sh 'go vet ./... &> reports/vet.txt'
                sh 'golint $(go list -f "{{.Dir}}" ./...) &> reports/lint.txt'
                stash includes: 'reports/*', name: 'unitStage'
            } 
            post {
                always {
                    archiveArtifacts 'reports/*'
                    junit 'reports/test.xml'
                }
            }
        }

        stage('Linting & static analyis') {
            parallel {
                stage('SonarQube') {
                    agent {
                        docker {
                            image 'bitflowstream/golang-build:debian'
                            args '-v /tmp/go-mod-cache/debian:/go'
                        }
                    }
                    steps {
                        dir('reports') {
                            unstash 'unitStage'
                        }
                        script {
                            def scannerHome = tool 'sonar-scanner-linux'
                            withSonarQubeEnv('CIT SonarQube') {
                                sh """
                                    ${scannerHome}/bin/sonar-scanner -Dsonar.projectKey=go-bitflow -Dsonar.branch.name=$BRANCH_NAME \
                                        -Dsonar.sources=. -Dsonar.tests=. \
                                        -Dsonar.inclusions="**/*.go" -Dsonar.test.inclusions="**/*_test.go" \
                                        -Dsonar.exclusions="script/script/internal/*.go" \
                                        -Dsonar.go.golint.reportPath=reports/lint.txt \
                                        -Dsonar.go.govet.reportPaths=reports/vet.txt \
                                        -Dsonar.go.coverage.reportPaths=reports/test-coverage.txt \
                                        -Dsonar.test.reportPath=reports/test.xml
                                """
                            }
                        }
                        timeout(time: 10, unit: 'MINUTES') {
                            waitForQualityGate abortPipeline: true
                        }
                    }
                }

                stage ("Lint Dockerfiles") {
                    agent {
                        docker {
                            image 'hadolint/hadolint:latest-debian'
                        }
                    }
                    steps {
                        sh 'hadolint --format checkstyle build/*.Dockerfile | tee -a hadolint.xml'
                    }
                    post {
                        always {
                            archiveArtifacts 'hadolint.xml'
                        }
                    }
                }
            }
        }

        stage('Build docker images') {
            parallel {
                stage('amd64') {
                    agent {
                        docker {
                            image 'bitflowstream/golang-build:alpine'
                            args '-v /tmp/go-mod-cache/alpine:/go -v /var/run/docker.sock:/var/run/docker.sock'
                        }
                    }
                    steps {
                        sh './build/native-build.sh'
                        script {
                            def image = docker.build registry + ":$BRANCH_NAME-build-$BUILD_NUMBER", '-f build/alpine-prebuilt.Dockerfile build/_output'
                        }
                        sh "./build/test-image.sh $BRANCH_NAME-build-$BUILD_NUMBER"

                        sh './build/native-static-build.sh'
                        script {
                            def staticImage = docker.build registry + ":static-$BRANCH_NAME-build-$BUILD_NUMBER",  '-f build/static-prebuilt.Dockerfile build/_output/static'
                        }
                        sh "./build/test-image.sh static-$BRANCH_NAME-build-$BUILD_NUMBER"
                    }
                    post {
                        always {
                            sh 'mv build/_output/bitflow-pipeline build/_output/bitflow-pipeline-amd64'
                            archiveArtifacts 'build/_output/bitflow-pipeline-amd64'
                        }
                        success {
                            script {
                                if (env.BRANCH_NAME == 'master') {
                                    docker.withRegistry('', registryCredential) {
                                        image.push("build-$BUILD_NUMBER")
                                        image.push("latest-amd64")
                                        staticImage.push("static-build-$BUILD_NUMBER")
                                        staticImage.push("static")
                                    }
                                }
                            }                                      
                        }
                    }
                }

                stage('arm32v7') {
                    agent {
                        docker {
                            image 'bitflowstream/golang-build:arm32v7'
                            args '-v /tmp/go-mod-cache/alpine:/go -v /var/run/docker.sock:/var/run/docker.sock'
                        }
                    }
                    steps {
                        sh './build/native-build.sh'
                        script {
                            def image = docker.build registry + ":$BRANCH_NAME-build-$BUILD_NUMBER-arm32v7", '-f build/arm32v7-prebuilt.Dockerfile build/_output'
                        }
                        sh "./build/test-image.sh $BRANCH_NAME-build-$BUILD_NUMBER-arm32v7"
                    }
                    post {
                        always {
                            sh 'mv build/_output/bitflow-pipeline build/_output/bitflow-pipeline-arm32v7'
                            archiveArtifacts 'build/_output/bitflow-pipeline-arm32v7'
                        }
                        success {
                            script {
                                if (env.BRANCH_NAME == 'master') {
                                    docker.withRegistry('', registryCredential) {
                                        image.push("build-$BUILD_NUMBER-arm32v7")
                                        image.push("latest-arm32v7")
                                    }
                                }
                            }                                      
                        }
                    }
                }

                stage('arm32v7 static') {
                    agent {
                        docker {
                            image 'bitflowstream/golang-build:static-arm32v7'
                            args '-v /tmp/go-mod-cache/debian:/go -v /var/run/docker.sock:/var/run/docker.sock'
                        }
                    }
                    steps {
                        sh './build/native-static-build.sh'
                        script {
                            def image = docker.build registry + ":static-$BRANCH_NAME-build-$BUILD_NUMBER-arm32v7", '-f build/static-prebuilt.Dockerfile build/_output/static'
                        }
                        sh "./build/test-image.sh static-$BRANCH_NAME-build-$BUILD_NUMBER-arm32v7"
                    }
                    post {
                        always {
                            sh 'mv build/_output/static/bitflow-pipeline build/_output/static/bitflow-pipeline-arm32v7'
                            archiveArtifacts 'build/_output/static/bitflow-pipeline-arm32v7'
                        }
                        success {
                            script {
                                if (env.BRANCH_NAME == 'master') {
                                    docker.withRegistry('', registryCredential) {
                                        image.push("static-build-$BUILD_NUMBER-arm32v7")
                                        image.push("static-arm32v7")
                                    }
                                }
                            }                                      
                        }
                    }
                }

                stage('arm64v8') {
                    agent {
                        docker {
                            image 'bitflowstream/golang-build:arm64v8'
                            args '-v /tmp/go-mod-cache/alpine:/go -v /var/run/docker.sock:/var/run/docker.sock'
                        }
                    }
                    steps {
                        sh './build/native-build.sh'
                        script {
                            def image = docker.build registry + ":$BRANCH_NAME-build-$BUILD_NUMBER-arm64v8", '-f build/arm64v8-prebuilt.Dockerfile build/_output'
                        }
                        sh "./build/test-image.sh $BRANCH_NAME-build-$BUILD_NUMBER-arm64v8"
                    }
                    post {
                        always {
                            sh 'mv build/_output/bitflow-pipeline build/_output/bitflow-pipeline-arm64v8'
                            archiveArtifacts 'build/_output/bitflow-pipeline-arm64v8'
                        }
                        success {
                            script {
                                if (env.BRANCH_NAME == 'master') {
                                    docker.withRegistry('', registryCredential) {
                                        image.push("build-$BUILD_NUMBER-arm64v8")
                                        image.push("latest-arm64v8")
                                    }
                                }
                            }                                      
                        }
                    }          
                }

                stage('arm64v8 static') {
                    agent {
                        docker {
                            image 'bitflowstream/golang-build:static-arm64v8'
                            args '-v /tmp/go-mod-cache/debian:/go -v /var/run/docker.sock:/var/run/docker.sock'
                        }
                    }
                    steps {
                        sh './build/native-static-build.sh'
                        script {
                            def image = docker.build registry + ":static-$BRANCH_NAME-build-$BUILD_NUMBER-arm64v8", '-f build/static-prebuilt.Dockerfile build/_output/static'
                        }
                        sh "./build/test-image.sh static-$BRANCH_NAME-build-$BUILD_NUMBER-arm64v8"
                    }
                    post {
                        always {
                            sh 'mv build/_output/static/bitflow-pipeline build/_output/static/bitflow-pipeline-arm64v8'
                            archiveArtifacts 'build/_output/static/bitflow-pipeline-arm64v8'
                        }
                        success {
                            script {
                                if (env.BRANCH_NAME == 'master') {
                                    docker.withRegistry('', registryCredential) {
                                        image.push("static-build-$BUILD_NUMBER-arm64v8")
                                        image.push("static-arm64v8")
                                    }
                                }
                            }                                      
                        }
                    }
                }
            }   
        }
        
        stage('Docker manifest') {
            when {
                branch 'master'
            }
            agent {
                docker {
                    image 'bitflowstream/golang-build:debian'
                    args '-v /var/run/docker.sock:/var/run/docker.sock'
                }
            }
            steps {
                withCredentials([[
                    $class: 'UsernamePasswordMultiBinding',
                    credentialsId: 'dockerhub', usernameVariable: 'DOCKERUSER', passwordVariable: 'DOCKERPASS'
                ]]) {
                    // Dockerhub Login
                    sh 'echo "$DOCKERPASS" | docker login -u "$DOCKERUSER" --password-stdin'
                    sh "docker manifest create ${registry}:latest ${registry}:latest-amd64 ${registry}:latest-arm32v7 ${registry}:latest-arm64v8"
                    sh "docker manifest annotate ${registry}:latest ${registry}:latest-arm32v7 --os=linux --arch=arm --variant=v7"
                    sh "docker manifest annotate ${registry}:latest ${registry}:latest-arm64v8 --os=linux --arch=arm64 --variant=v8"
                    sh "docker manifest push --purge ${registry}:latest"
                }
            }
        }
    }

    post {
        success {
            node('master') {
                withSonarQubeEnv('CIT SonarQube') {
                    slackSend channel: '#jenkins-builds-all', color: 'good',
                        message: "Build ${env.JOB_NAME} ${env.BUILD_NUMBER} was successful (<${env.BUILD_URL}|Open Jenkins>) (<${env.SONAR_HOST_URL}|Open SonarQube>)"
                }
            }
        }
        failure {
            node('master') {
                slackSend channel: '#jenkins-builds-all', color: 'danger',
                    message: "Build ${env.JOB_NAME} ${env.BUILD_NUMBER} failed (<${env.BUILD_URL}|Open Jenkins>)"
            }
        }
        fixed {
            node('master') {
                withSonarQubeEnv('CIT SonarQube') {
                    slackSend channel: '#jenkins-builds', color: 'good',
                        message: "Thanks to ${env.GIT_COMMITTER_EMAIL}, build ${env.JOB_NAME} ${env.BUILD_NUMBER} was successful (<${env.BUILD_URL}|Open Jenkins>) (<${env.SONAR_HOST_URL}|Open SonarQube>)"
                }
            }
        }
        regression {
            node('master') {
                slackSend channel: '#jenkins-builds', color: 'danger',
                    message: "What have you done ${env.GIT_COMMITTER_EMAIL}? Build ${env.JOB_NAME} ${env.BUILD_NUMBER} failed (<${env.BUILD_URL}|Open Jenkins>)"
            }
        }
    }
}
