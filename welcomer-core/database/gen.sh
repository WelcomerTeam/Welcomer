sqlc generate
find . -type f -exec sed -i 's/"github\.com\/google\/uuid"//g' {} +
