# Render.com Deployment Checklist

## Pre-Deployment
- [ ] Code is working locally
- [ ] MongoDB Atlas is set up with connection string
- [ ] All dependencies are in `go.mod` and `go.sum`
- [ ] `.env` is in `.gitignore` (✅ Already done)
- [ ] Health check endpoint works: `/api/health`

## GitHub Setup
- [ ] Create GitHub repository
- [ ] Push code to GitHub
  ```bash
  git init
  git add .
  git commit -m "Initial commit for Render deployment"
  git remote add origin https://github.com/YOUR_USERNAME/flowkit-backend.git
  git branch -M main
  git push -u origin main
  ```

## Render.com Setup
- [ ] Create Render account: https://render.com
- [ ] Connect GitHub account
- [ ] Create new Web Service
- [ ] Select your repository
- [ ] Choose root directory (if needed)

## Service Configuration
- [ ] Name: `flowkit-api`
- [ ] Region: Choose closest to users
- [ ] Branch: `main`
- [ ] Runtime: Go
- [ ] Build Command: `go build -o flowkit-api main.go`
- [ ] Start Command: `./flowkit-api`
- [ ] Plan: Free

## Environment Variables
Add these in Render dashboard:

- [ ] `MONGODB_URI` = `mongodb+srv://rachealaudu_db_user:PASSWORD@cluster0.gr27rks.mongodb.net/flowkit_leave_management`
- [ ] `JWT_SECRET` = (Generate secure random string or use Render's generator)
- [ ] `GIN_MODE` = `release`
- [ ] `PORT` = `5000`

## MongoDB Configuration
- [ ] Login to MongoDB Atlas
- [ ] Go to Security → Network Access
- [ ] Add IP Address → Allow access from anywhere (0.0.0.0/0)
- [ ] Confirm changes

## Deploy
- [ ] Click "Create Web Service" on Render
- [ ] Wait for deployment to complete (watch logs)
- [ ] Check deployment status shows "Live"
- [ ] Copy your Render URL (e.g., `https://flowkit-api-xxxx.onrender.com`)

## Testing
- [ ] Visit health endpoint: `https://your-app.onrender.com/api/health`
- [ ] Test login endpoint with curl or Postman
- [ ] Verify database connection in logs
- [ ] Test a few key endpoints

## Frontend Update
- [ ] Update API base URL in frontend code
  - From: `http://localhost:5000/api`
  - To: `https://your-app.onrender.com/api`
- [ ] Test frontend with new backend URL
- [ ] Verify all features work (login, leaves, dashboard)

## Post-Deployment
- [ ] Monitor logs for any errors
- [ ] Set up email notifications for deployment status (optional)
- [ ] Document your Render URL for team
- [ ] Test on different devices/networks

## Optional Enhancements
- [ ] Add custom domain
- [ ] Set up SSL certificate (automatic on Render)
- [ ] Enable auto-deploy on GitHub push (enabled by default)
- [ ] Upgrade to paid plan if needed ($7/month for always-on)

## Troubleshooting
If deployment fails:
- [ ] Check build logs in Render dashboard
- [ ] Verify all environment variables are set correctly
- [ ] Ensure MongoDB allows Render's IP addresses
- [ ] Check Go version compatibility (1.21+)
- [ ] Verify `go.mod` and `go.sum` are committed

## Your Deployed URLs
Backend: `https://________________________________.onrender.com`
Health Check: `https://________________________________.onrender.com/api/health`

## Notes
- Free tier sleeps after 15 min inactivity (wakes in ~30 sec on first request)
- Render auto-deploys on every git push to main branch
- Logs are available in Render dashboard
- MongoDB Atlas connection string should use DATABASE_NAME from URI

---
**Status:** ⬜ Not Started | 🔄 In Progress | ✅ Completed | ❌ Failed
