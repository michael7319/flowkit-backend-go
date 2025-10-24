# 🚀 FlowKit Backend - Ready for Render Deployment!

## ✅ What's Already Done

Your backend is **deployment-ready**! Here's what you already have:

1. **Dockerfile** - For containerized deployment ✅
2. **render.yaml** - Render configuration file ✅
3. **.gitignore** - Protects sensitive files ✅
4. **Environment variable support** - PORT, GIN_MODE, etc. ✅
5. **CORS enabled** - Frontend can connect ✅
6. **Security headers** - Production-ready ✅
7. **MongoDB connection** - Already working ✅

## 📋 Quick Start Guide

### 1️⃣ Push to GitHub (5 minutes)

```bash
# Navigate to backend folder
cd backend-go

# Initialize git (if not already)
git init

# Add all files
git add .

# Commit
git commit -m "Ready for Render deployment"

# Create repo on GitHub, then:
git remote add origin https://github.com/YOUR_USERNAME/flowkit-backend.git
git branch -M main
git push -u origin main
```

### 2️⃣ Deploy on Render (10 minutes)

1. **Go to Render.com**
   - Visit: https://render.com
   - Sign up/Login (free account)

2. **Create Web Service**
   - Click "New +" → "Web Service"
   - Connect your GitHub
   - Select `flowkit-backend` repo
   - Render will auto-detect Go

3. **Configure (auto-filled from render.yaml)**
   - Name: `flowkit-api`
   - Runtime: Go
   - Build: `go build -o flowkit-api main.go`
   - Start: `./flowkit-api`
   - Plan: Free

4. **Add Environment Variables**
   Click "Environment" tab:
   
   | Variable | Value |
   |----------|-------|
   | `MONGODB_URI` | Your MongoDB Atlas connection string |
   | `JWT_SECRET` | Click "Generate" button |
   | `GIN_MODE` | `release` |
   | `PORT` | `5000` |

5. **Deploy!**
   - Click "Create Web Service"
   - Wait ~2-3 minutes for build
   - Get your URL: `https://flowkit-api-xxxx.onrender.com`

### 3️⃣ Update Frontend (2 minutes)

Find your API configuration (likely in `src/services/api.js`):

```javascript
// Change this:
const API_BASE_URL = 'http://localhost:5000/api';

// To this (use your actual Render URL):
const API_BASE_URL = 'https://flowkit-api-xxxx.onrender.com/api';
```

### 4️⃣ Test (2 minutes)

Visit: `https://your-app.onrender.com/api/health`

Should see:
```json
{
  "status": "healthy",
  "message": "FlowKit Leave Management API is running"
}
```

## 🔑 Important Information

### Your MongoDB Connection String
```
mongodb+srv://rachealaudu_db_user:YOUR_PASSWORD@cluster0.gr27rks.mongodb.net/flowkit_leave_management
```
**⚠️ Replace `YOUR_PASSWORD` with your actual MongoDB password!**

### MongoDB Atlas Setup
1. Go to MongoDB Atlas → Security → Network Access
2. Click "Add IP Address"
3. Select "Allow access from anywhere" (0.0.0.0/0)
4. Click "Confirm"

### Free Tier Limitations
- ✅ 750 hours/month (enough for always-on)
- ⚠️ Sleeps after 15 min inactivity
- ✅ Wakes up in ~30 seconds on first request
- ✅ Perfect for development/testing
- 💰 Upgrade to $7/month for always-on

## 📁 Files Created

- `render.yaml` - Render configuration
- `RENDER_DEPLOYMENT.md` - Detailed deployment guide
- `DEPLOYMENT_CHECKLIST.md` - Step-by-step checklist

## 🆘 Quick Troubleshooting

**Build Failed?**
- Check Render logs tab
- Ensure `go.mod` and `go.sum` are committed

**Database Connection Failed?**
- Verify MongoDB URI in environment variables
- Check MongoDB Network Access allows 0.0.0.0/0
- Confirm password in connection string

**CORS Errors?**
- Already configured for all origins ✅
- For production, update to specific domain

**Service Unavailable?**
- Free tier sleeps after 15 min
- First request wakes it up (~30 sec)

## 🎯 Next Steps

1. ✅ Review `DEPLOYMENT_CHECKLIST.md`
2. ✅ Follow `RENDER_DEPLOYMENT.md` for detailed instructions
3. ✅ Push to GitHub
4. ✅ Deploy on Render
5. ✅ Update frontend URL
6. 🎉 You're live!

## 📞 Support Resources

- **Render Docs:** https://render.com/docs
- **Render Community:** https://community.render.com
- **Go on Render:** https://render.com/docs/deploy-go

## 🌟 Pro Tips

1. **Auto-Deploy:** Render auto-deploys on every push to main
2. **Logs:** Always available in Render dashboard
3. **Environment Variables:** Change anytime without redeploying
4. **Custom Domain:** Add your own domain in Settings
5. **Monitoring:** Set up health check alerts

---

**Ready to deploy?** Just push to GitHub and create the web service on Render!

**Estimated Total Time:** ~20 minutes for first deployment

**Your backend will be live at:** `https://flowkit-api-xxxx.onrender.com`

🚀 Good luck with your deployment!
