# SETUP REQUIRED: Add Animation GIFs

This directory needs actual GIF animation files for the application to run.

## Required Files

Create these GIF animation files in this directory:

- `idle.gif` - Default character animation (looping)
- `talking.gif` - Character speaking animation
- `happy.gif` - Happy/excited character animation  
- `sad.gif` - Sad/disappointed character animation

## Quick Setup for Testing

### Option 1: Use Sample GIFs
Download sample animated GIFs from:
- https://tenor.com (search for "pixel character animated")
- https://giphy.com (search for "8bit character loop")
- Create simple GIFs using online tools like ezgif.com

### Option 2: Create Simple Test GIFs
1. Use any image editor to create a 64x64 pixel image
2. Create 2-3 frames with slight variations
3. Export as animated GIF with transparency
4. Copy to this directory with the correct filenames

## File Requirements

- **Format**: Animated GIF with transparency
- **Size**: 64x64 to 256x256 pixels recommended
- **Frames**: 2-10 frames per animation
- **Timing**: 100-200ms per frame for smooth animation
- **File Size**: Keep under 1MB each

## Testing the Application

Once you have the GIF files:

```bash
# From the project root directory
go run cmd/companion/main.go -debug
```

The application will show errors if GIF files are missing or invalid.

## Example Directory Structure

```
animations/
├── idle.gif      (required)
├── talking.gif   (required)  
├── happy.gif     (required)
├── sad.gif       (required)
└── README.md     (this file)
```

## Creating Your Own Character

1. Design your character in your preferred art style
2. Create animation frames for each required state
3. Use consistent character positioning across all animations
4. Export as GIFs with transparency backgrounds
5. Test animations in the application
6. Adjust timing and file sizes as needed

The application will automatically loop animations and switch between them based on user interactions.
