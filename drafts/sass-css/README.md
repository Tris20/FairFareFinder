# setup
Youtube setup video
https://www.youtube.com/watch?v=wYWf2m_yzBQ&list=PL4-IK0AVhVjMYRhK9vRPatSlb-9r0aKgh

Below are key instructions
## global
1. install npm
  - https://nodejs.org/en/download
2. install sass
  - npm install -g sass
## project specific
3. setup npm in the project 
  - npm init -y
4. install sass to the project
  - npm install sass --save-dev

# use
sass --watch src/scss:dist/css

This means any time a sass file changes in the src/scss folder, the respective .css file will be created and updated in the dist/css folder. You can name the output folder however you like, so we could have just --watch src/scss:src/css to put this in a css folder


## Check this works
- in a terminal, run the watch command on a folder or file such as above, then change the value of a $variable. For example change $red: red; to $red:blue; and when saved, the main.css file should mirror those changes
