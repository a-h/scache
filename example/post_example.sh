curl -d '{ "id": "123", "name": "Test Person", "email": "test@example.com" }' https://dfcc3bjmul.execute-api.us-east-1.amazonaws.com/dev/user/123
curl -d '{ "id": "123", "name": "Test Person Updated", "email": "test@example.com" }' https://dfcc3bjmul.execute-api.us-east-1.amazonaws.com/dev/user/123
curl https://dfcc3bjmul.execute-api.us-east-1.amazonaws.com/dev/user/123