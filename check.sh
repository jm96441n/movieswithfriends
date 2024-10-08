GITHUB_OUTPUT=""
files=$(git grep -n '\bbinding\.pry\b' -- ':!*.md' ':!sorbet/**')
if [[ -n $files ]]; then
  {
    echo 'JSON_RESPONSE<<EOF'
    curl https://example.com
    echo EOF
  } >>./out.txt
fi
