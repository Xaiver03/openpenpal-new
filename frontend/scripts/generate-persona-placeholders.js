#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

// Define personas
const personas = [
  { id: 'poet', name: '诗人', color: '#8B5CF6' },
  { id: 'philosopher', name: '哲学家', color: '#3B82F6' },
  { id: 'artist', name: '艺术家', color: '#F59E0B' },
  { id: 'scientist', name: '科学家', color: '#10B981' },
  { id: 'traveler', name: '旅行者', color: '#EF4444' },
  { id: 'historian', name: '历史学家', color: '#6366F1' },
  { id: 'mentor', name: '人生导师', color: '#14B8A6' },
  { id: 'friend', name: '知心朋友', color: '#EC4899' }
];

// Create SVG placeholder for each persona
personas.forEach(persona => {
  const svg = `<?xml version="1.0" encoding="UTF-8"?>
<svg width="200" height="200" viewBox="0 0 200 200" xmlns="http://www.w3.org/2000/svg">
  <!-- Background Circle -->
  <circle cx="100" cy="100" r="90" fill="${persona.color}" opacity="0.1"/>
  
  <!-- Main Circle -->
  <circle cx="100" cy="100" r="80" fill="${persona.color}" opacity="0.2" stroke="${persona.color}" stroke-width="2"/>
  
  <!-- Inner Circle -->
  <circle cx="100" cy="100" r="60" fill="none" stroke="${persona.color}" stroke-width="1" opacity="0.5"/>
  
  <!-- Text Initial -->
  <text x="100" y="100" text-anchor="middle" dominant-baseline="central" 
        font-family="system-ui, -apple-system, sans-serif" 
        font-size="48" font-weight="bold" fill="${persona.color}">
    ${persona.name.charAt(0)}
  </text>
  
  <!-- Decorative Elements -->
  <circle cx="40" cy="40" r="3" fill="${persona.color}" opacity="0.5"/>
  <circle cx="160" cy="40" r="3" fill="${persona.color}" opacity="0.5"/>
  <circle cx="40" cy="160" r="3" fill="${persona.color}" opacity="0.5"/>
  <circle cx="160" cy="160" r="3" fill="${persona.color}" opacity="0.5"/>
</svg>`;

  const outputPath = path.join(__dirname, `../public/images/personas/${persona.id}.svg`);
  fs.writeFileSync(outputPath, svg);
  console.log(`Created ${outputPath}`);
});

// Also create PNG versions using a simple HTML canvas approach
const pngScript = `
<!DOCTYPE html>
<html>
<head>
  <script>
    function createPNG(id, name, color) {
      const canvas = document.createElement('canvas');
      canvas.width = 200;
      canvas.height = 200;
      const ctx = canvas.getContext('2d');
      
      // Background
      ctx.fillStyle = color + '20';
      ctx.beginPath();
      ctx.arc(100, 100, 90, 0, 2 * Math.PI);
      ctx.fill();
      
      // Main circle
      ctx.strokeStyle = color;
      ctx.lineWidth = 2;
      ctx.beginPath();
      ctx.arc(100, 100, 80, 0, 2 * Math.PI);
      ctx.stroke();
      
      // Text
      ctx.fillStyle = color;
      ctx.font = 'bold 48px system-ui';
      ctx.textAlign = 'center';
      ctx.textBaseline = 'middle';
      ctx.fillText(name.charAt(0), 100, 100);
      
      return canvas.toDataURL('image/png');
    }
    
    // Log the data URLs for manual saving
    const personas = ${JSON.stringify(personas)};
    personas.forEach(p => {
      console.log(p.id + '.png:', createPNG(p.id, p.name, p.color));
    });
  </script>
</head>
<body>
  <p>Open console to see PNG data URLs</p>
</body>
</html>
`;

fs.writeFileSync(path.join(__dirname, '../public/images/personas/generate-png.html'), pngScript);

console.log('\nSVG placeholders created successfully!');
console.log('To generate PNG versions, open public/images/personas/generate-png.html in a browser.');