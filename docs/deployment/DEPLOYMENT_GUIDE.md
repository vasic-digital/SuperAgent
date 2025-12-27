# SuperAgent Documentation Deployment Guide

## Quick Start

### GitHub Pages Deployment (Recommended)
1. **Enable GitHub Pages** in repository settings
2. **Push documentation** to `main` branch
3. **Access docs** at `https://[username].github.io/[repository]/`

### Local Development
```bash
cd Website
npm install
npm run dev  # Serves at http://localhost:8080
```

## Documentation Structure

```
docs/                    # Markdown documentation
├── README.md           # Main documentation index
├── api/                # API documentation
├── developer/          # Developer guides
├── user/               # User guides
└── deployment/         # Deployment guides

Website/                 # Website files
├── public/index.html   # Main website
├── styles/main.css     # Website styles
├── scripts/main.js     # Website JavaScript
└── package.json        # Dependencies
```

## Deployment Options

### 1. GitHub Pages (Free)
- **Setup**: Enable in repository settings
- **Deploy**: Push to main branch
- **URL**: `https://[username].github.io/[repo]/`
- **Best for**: Open source projects

### 2. Netlify (Free tier available)
- **Setup**: Connect GitHub repository
- **Build**: `cd Website && npm run build`
- **Publish**: `Website/public`
- **Best for**: Custom domains, forms

### 3. Vercel (Free tier available)
- **Setup**: Import GitHub repository
- **Build**: `cd Website && npm run build`
- **Publish**: `Website/public`
- **Best for**: Performance, analytics

### 4. Docker (Self-hosted)
```bash
docker build -t superagent-docs .
docker run -d -p 8080:80 superagent-docs
```

## Customization

### Website Colors
Edit `Website/styles/main.css`:
```css
:root {
  --primary-color: #2563eb;
  --secondary-color: #10b981;
}
```

### Content Updates
- **Documentation**: Edit `.md` files in `docs/`
- **Website**: Edit `Website/public/index.html`
- **API docs**: Update `docs/api/openapi.yaml`

## Performance Tips

1. **Optimize images** before adding to assets
2. **Enable compression** on your hosting platform
3. **Use CDN** for static assets
4. **Minimize CSS/JS** in production builds

## Support

- **Issues**: GitHub Issues
- **Documentation**: `docs/README.md`
- **Examples**: See existing documentation files

---

*Last updated: December 2025*