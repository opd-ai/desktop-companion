# Animation Assets

This directory contains GIF animation files for the default character.

## Required Animations

The character configuration requires these animations:

- `idle.gif` - Default animation when character is not interacting
- `talking.gif` - Animation played when character is speaking
- `happy.gif` - Animation for positive interactions  
- `sad.gif` - Animation for negative or error states

## File Requirements

- **Format**: Animated GIF with transparency support
- **Size**: Recommended 64x64 to 256x256 pixels
- **File Size**: Keep under 1MB each for best performance
- **Frame Rate**: 10-15 FPS recommended
- **Colors**: Use indexed color mode for smaller file sizes

## Creating Custom Animations

1. Create your animation frames as individual images
2. Use a tool like GIMP, Photoshop, or online GIF makers to combine frames
3. Ensure the GIF has a transparent background for desktop overlay
4. Test the animation by updating the character.json file

## Placeholder Animations

The current animations are placeholders. To add real animations:

1. Replace the existing GIF files with your custom animations
2. Ensure filenames match the character.json configuration
3. Restart the application to load new animations

## Tools for Creating GIFs

- **GIMP**: Free, supports transparency and frame timing
- **Photoshop**: Professional tool with excellent GIF export
- **Online**: ezgif.com, giphy.com (for simple animations)
- **Command Line**: ffmpeg, ImageMagick for batch processing
