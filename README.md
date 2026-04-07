# RealWorld-Vue-Golang-SocialApp

## Environment Variables

Create or update `backend/api/.env` with the values below:

```env
JWT_SECRET=your-secret-key

# Default page size for post listings when `limit` is not provided.
POSTS_PAGE_SIZE=2

# Maximum allowed `limit` for post listings.
POSTS_MAX_PAGE_SIZE=50
```

`POSTS_PAGE_SIZE` controls the default number of posts returned per page in `GetAllPosts`.
`POSTS_MAX_PAGE_SIZE` caps the client-provided `limit` query parameter.
