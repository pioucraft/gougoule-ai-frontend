# start by getting commit message
echo "Enter commit message:"
read commit_message

git add .
git commit -m "$commit_message"
git push

vc commit -m "$commit_message"
