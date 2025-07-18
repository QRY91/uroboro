# uroboro.dev Landing Page

The official landing page for uroboro - Automatic development work capture with timeline visualization.

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

- **Timeline Visualization Showcase** - Interactive demonstration of development journey visualization
- **Smart Feature Demonstration** - Examples of automatic project detection and auto-tagging
- **Responsive Design** - Works on desktop, tablet, and mobile
- **Interactive Voice Demo** - Click between writing styles to see examples
- **Smooth Animations** - Subtle scroll animations and transitions
- **Developer-Focused** - Clean, technical aesthetic that resonates with developers
- **Performance Optimized** - Minimal dependencies, fast loading

## 🎯 Sections

1. **Hero** - Development work capture with timeline visualization value proposition
2. **Timeline Showcase** - Interactive demonstration of the killer feature
3. **Problem** - Resonates with developer content creation struggles  
4. **Solution** - Simple 3-command workflow with automatic capture
5. **Demo Carousel** - Interactive demonstrations of core timeline workflow
6. **Truth Section** - Benefits of timeline visualization and three-command philosophy
7. **Technical Deep-dive** - Local-first benefits and privacy focus
8. **CTA** - Clear next steps for visitors

## 🎬 Key Features Highlighted

- **Timeline Visualization** - See your development work as a beautiful interactive story
- **Smart Project Detection** - Zero-configuration auto-discovery from git repos and package files
- **Content-Based Auto-Tagging** - Automatic pattern analysis for development insights
- **Local-First Privacy** - Your data stays on your machine, no cloud dependencies

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
- Hero copy: Edit `index.html` hero section (now includes Trinity messaging)
- Trinity features: Edit Trinity integration section in `index.html`
- Voice examples: Edit `voiceExamples` object in `script.js`
- Contact info: Update footer links in `index.html`

### Adding New Content Examples
1. Add timeline examples to showcase different project types
2. Add content generation examples to `voiceExamples` in `script.js`
3. Add new demonstration sections as needed

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

This landing page showcases uroboro's timeline visualization and development capture capabilities. To contribute:

1. Fork the main repository
2. Make changes in the `landing-page/` directory  
3. Test locally with `npm run dev` or `python3 -m http.server 8000`
4. Ensure messaging focuses on timeline visualization as the killer feature
5. Submit a pull request

## 📄 License

MIT License - see the main uroboro repository for details. 