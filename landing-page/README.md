# uroboro.dev Landing Page

The official landing page for uroboro - The Self-Documenting Content Pipeline.

## 🚀 Quick Start

```bash
# Clone the repository
git clone https://github.com/qry91/uroboro
cd uroboro/landing-page

# Start development server
npm run dev
# or
python3 -m http.server 8000

# Open http://localhost:8000 in your browser
```

## 📁 Structure

```
landing-page/
├── index.html          # Main landing page
├── style.css           # Modern, responsive styling  
├── script.js           # Interactive voice demo + animations
├── package.json        # Project configuration
└── README.md           # This file
```

## 🎨 Features

- **Responsive Design** - Works on desktop, tablet, and mobile
- **Interactive Voice Demo** - Click between writing styles to see examples
- **Smooth Animations** - Subtle scroll animations and transitions
- **Developer-Focused** - Clean, technical aesthetic that resonates with developers
- **Performance Optimized** - Minimal dependencies, fast loading

## 🎯 Sections

1. **Hero** - Clear value proposition with quick demo
2. **Problem** - Resonates with developer content creation struggles  
3. **Solution** - 3-step process explanation with architecture
4. **Voice Showcase** - Interactive demo of 5 writing styles
5. **Technical Deep-dive** - Local AI benefits and requirements
6. **CTA** - Clear next steps for visitors

## 🎪 Voice Styles Demonstrated

- **Professional Conversational** - Approachable but polished
- **Technical Deep-dive** - Detailed technical explanations
- **Storytelling** - Narrative approach to development stories
- **Minimalist** - Concise, bullet-focused content
- **Thought Leadership** - Industry insights and bigger picture

## 🚀 Deployment Options

### Vercel (Recommended)
```bash
# Install Vercel CLI
npm i -g vercel

# Deploy
vercel --prod
```

### Netlify
```bash
# Install Netlify CLI  
npm i -g netlify-cli

# Deploy
netlify deploy --prod --dir .
```

### GitHub Pages
```bash
# Push to gh-pages branch
git subtree push --prefix landing-page origin gh-pages
```

### Traditional Hosting
Upload all files to your web server. The site is pure HTML/CSS/JS with no build step required.

## 🔧 Customization

### Colors
Edit CSS variables in `style.css`:
```css
:root {
    --primary: #2563eb;        /* Main brand color */
    --accent: #06b6d4;         /* Accent color */  
    --bg-primary: #ffffff;     /* Background */
    /* ... */
}
```

### Content
- Hero copy: Edit `index.html` hero section
- Voice examples: Edit `voiceExamples` object in `script.js`
- Contact info: Update footer links in `index.html`

### Adding New Voices
1. Add voice configuration to uroboro `config/settings.json`
2. Add example to `voiceExamples` in `script.js`
3. Add tab button to voice showcase section

## 📊 Analytics

Add your analytics code before the closing `</body>` tag in `index.html`:

```html
<!-- Google Analytics -->
<script async src="https://www.googletagmanager.com/gtag/js?id=GA_MEASUREMENT_ID"></script>
<script>
  window.dataLayer = window.dataLayer || [];
  function gtag(){dataLayer.push(arguments);}
  gtag('js', new Date());
  gtag('config', 'GA_MEASUREMENT_ID');
</script>

<!-- Plausible Analytics (privacy-friendly alternative) -->
<script defer data-domain="uroboro.dev" src="https://plausible.io/js/script.js"></script>
```

## 🎯 SEO Optimization

The page includes:
- Semantic HTML structure
- Meta descriptions and Open Graph tags  
- Proper heading hierarchy (h1, h2, h3)
- Alt text for important visual elements
- Fast loading performance
- Mobile-responsive design

To enhance SEO:
1. Add more specific meta descriptions
2. Include structured data (JSON-LD)
3. Optimize images with next-gen formats
4. Add blog content for long-tail keywords

## 🤝 Contributing

This landing page is part of the main uroboro project. To contribute:

1. Fork the main repository
2. Make changes in the `landing-page/` directory
3. Test locally with `npm run dev`
4. Submit a pull request

## 📄 License

MIT License - see the main uroboro repository for details. 