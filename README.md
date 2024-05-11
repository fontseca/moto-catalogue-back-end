# Motorcycle Catalogue API

This is a non-realistic back-end side project for a motorcycle catalogue mobile application.

## API Endpoints

| Actor | Method   | Endpoint                                    | Description                                                                        |
|-------|----------|---------------------------------------------|------------------------------------------------------------------------------------|
| Any   | `POST`   | `/signup`                                   | Register a new user.                                                               |
| Any   | `POST`   | `/login`                                    | Sign in a registered user.                                                         |
| User  | `GET`    | `/me`                                       | Get information about the authenticated user.                                      |
| User  | `PATCH`  | `/me`                                       | Partially update information about the authenticated user.                         |
| User  | `DELETE` | `/me`                                       | Delete the authenticated user's account and every objected related to him.         |
| User  | `POST`   | `/me/motorcycles`                           | Create a new motorcycle entry for the authenticated user.                          |
| User  | `GET`    | `/me/motorcycles`                           | Get a list of motorcycles owned by the authenticated user.                         |
| User  | `GET`    | `/me/motorcycles/{motorcycle_id}`           | Get details of a specific motorcycle owned by the authenticated user.              |
| User  | `PATCH`  | `/me/motorcycles/{motorcycle_id}`           | Partially update details of a specific motorcycle owned by the authenticated user. |
| User  | `DELETE` | `/me/motorcycles/{motorcycle_id}`           | Delete a motorcycle of the authenticated user.                                     |
| User  | `POST`   | `/me/motorcycles/favorites`                 | Add a motorcycle to the favorites list of the authenticated user.                  | 
| User  | `GET`    | `/me/motorcycles/favorites`                 | Get the favorite motorcycles of the authenticated user.                            | 
| User  | `DELETE` | `/me/motorcycles/favorites/{motorcycle_id}` | Remove a motorcycle from the favorites list of the authenticated user.             |
| User  | `GET`    | `/users`                                    | Get a list of all users.                                                           |
| User  | `GET`    | `/users/{user_id}`                          | Get details of a specific user.                                                    |
| User  | `GET`    | `/motorcycles`                              | Get a list of all motorcycles.                                                     |
| User  | `GET`    | `/motorcycles/{motorcycle_id}`              | Get details of a specific motorcycle with the owner's details.                     |
