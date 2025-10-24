# FlowKit Backend - Render.com Deployment Guide

## Prerequisites
- GitHub account
- Render.com account (free)
- MongoDB Atlas account (free) - You already have this!

## Step 1: Prepare Your Code

### 1.1 Push to GitHub
```bash
# Initialize git if not already done
cd backend-go
git init
git add .
git commit -m "Prepare for Render deployment"

# Create a new repository on GitHub, then:
git remote add origin https://github.com/YOUR_USERNAME/flowkit-backend.git
git branch -M main
git push -u origin main
```

## Step 2: Deploy on Render

### 2.1 Create New Web Service
1. Go to https://render.com and sign in
2. Click **"New +"** → **"Web Service"**
3. Connect your GitHub repository
4. Select the `flowkit-backend` repository
5. Render will auto-detect it's a Go app

### 2.2 Configure Service
Render should auto-fill most settings from `render.yaml`, but verify:

**Basic Settings:**
- **Name:** `flowkit-api` (or your choice)
- **Region:** Choose closest to your users (e.g., Oregon, Singapore)
- **Branch:** `main`
- **Root Directory:** Leave blank (or `backend-go` if repo has multiple folders)

**Build & Deploy:**
- **Runtime:** Go
- **Build Command:** `go build -o flowkit-api main.go`
- **Start Command:** `./flowkit-api`

**Plan:**
- Select **"Free"** plan
- ⚠️ Note: Free tier sleeps after 15 min inactivity (wakes up in ~30 seconds)

### 2.3 Set Environment Variables
Click **"Environment"** tab and add:

| Key | Value | Notes |
|-----|-------|-------|
| `MONGODB_URI` | `mongodb+srv://rachealaudu_db_user:YOUR_PASSWORD@cluster0.gr27rks.mongodb.net/flowkit_leave_management` | Your existing MongoDB Atlas connection string |
| `JWT_SECRET` | Click "Generate" or use your own secure string | Must be strong and random |
| `GIN_MODE` | `release` | Production mode |
| `PORT` | `5000` | Render provides PORT, but we set default |

**Important:** Replace `YOUR_PASSWORD` in MongoDB URI with your actual password!

### 2.4 Deploy
1. Click **"Create Web Service"**
2. Render will automatically:
   - Clone your repository
   - Install dependencies (`go mod download`)
   - Build the application
   - Start the server
   - Provide a URL like: `https://flowkit-api.onrender.com`

## Step 3: Update Frontend

After deployment, update your frontend API base URL:

**In `src/services/api.js` (or wherever you have it):**
```javascript
// Change from:
const API_BASE_URL = 'http://localhost:5000/api';

// To:
const API_BASE_URL = 'https://flowkit-api.onrender.com/api';
// Or use your actual Render URL
```

## Step 4: Test Your Deployment

### 4.1 Health Check
Visit: `https://your-app.onrender.com/api/health`

Should return:
```json
{
  "status": "healthy",
  "message": "FlowKit Leave Management API is running"
}
```

### 4.2 Test Login
```bash
curl -X POST https://your-app.onrender.com/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"your@email.com","password":"yourpassword"}'
```

## Step 5: Configure MongoDB Network Access

Make sure MongoDB Atlas allows connections from Render:

1. Go to MongoDB Atlas → Security → Network Access
2. Click "Add IP Address"
3. Select **"Allow access from anywhere"** (0.0.0.0/0)
   - Or add Render's IP ranges specifically
4. Click "Confirm"

## Step 6: Enable CORS (Already Configured!)

Your backend already has CORS enabled for all origins:
```go
corsConfig.AllowAllOrigins = true
```

For production, you might want to restrict this to your frontend domain:
```go
corsConfig.AllowOrigins = []string{
    "https://your-frontend-domain.com",
    "http://localhost:5173", // for local development
}
```

## Monitoring & Logs

### View Logs
1. Go to your service dashboard on Render
2. Click **"Logs"** tab
3. You'll see real-time logs including:
   - Database connections
   - API requests
   - Errors

### Restart Service
- Click **"Manual Deploy"** → **"Deploy latest commit"**
- Or use "Restart" to restart without rebuilding

## Auto-Deploy on Git Push

Render automatically deploys when you push to your GitHub branch:

```bash
git add .
git commit -m "Update feature"
git push origin main
# Render will auto-deploy!
```

## Troubleshooting

### Issue: Service won't start
**Check:**
1. Logs tab for error messages
2. MongoDB URI is correct
3. All environment variables are set
4. Build command completed successfully

### Issue: Database connection failed
**Check:**
1. MongoDB Atlas Network Access allows 0.0.0.0/0
2. MongoDB URI has correct password
3. Database name is correct in URI

### Issue: CORS errors
**Check:**
1. Frontend is using HTTPS for Render URL
2. CORS is enabled in backend (it already is!)

### Issue: 503 Service Unavailable
**Reason:** Free tier sleeps after 15 min inactivity
**Solution:** First request will wake it up (~30 seconds)
**Alternative:** Upgrade to paid plan ($7/month) for always-on

## Custom Domain (Optional)

To use your own domain:
1. Go to **"Settings"** → **"Custom Domain"**
2. Add your domain
3. Update DNS records as shown by Render

## Environment-Specific Tips

### Development
```bash
# Run locally
go run main.go
```

### Staging/Production
- Use `GIN_MODE=release`
- Set strong JWT_SECRET
- Monitor logs regularly
- Set up alerts in Render

## Cost Breakdown

**Free Tier:**
- 750 hours/month runtime
- Sleeps after 15 min inactivity
- 512 MB RAM
- Shared CPU
- Perfect for development/testing

**Starter ($7/month):**
- Always on
- 512 MB RAM
- Shared CPU
- Good for small production apps

## Next Steps

1. ✅ Push code to GitHub
2. ✅ Create Render account
3. ✅ Deploy service
4. ✅ Set environment variables
5. ✅ Update frontend URL
6. ✅ Test all endpoints
7. 🎉 Your backend is live!

## Support

- Render Docs: https://render.com/docs
- Render Community: https://community.render.com
- Go on Render: https://render.com/docs/deploy-go

## Your Backend URL

After deployment, your API will be available at:
```
https://flowkit-api-XXXX.onrender.com
```

All endpoints will be:
- `POST /api/auth/register`
- `POST /api/auth/login`
- `GET /api/leaves/my-leaves`
- etc.

---

**Note:** Remember to update your frontend's API base URL after deployment!
