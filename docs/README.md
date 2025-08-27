# TEE Auth Documentation

This folder contains the Mintlify documentation for the TEE Auth project.

## Local Development

To run the documentation locally, follow these steps:

1. Install the Mintlify CLI:

```bash
npm i -g mintlify
```

2. Run the development server:

```bash
mintlify dev
```

This will start a local server at http://localhost:3000.

## Structure

- `mint.json` - Main configuration file
- `introduction.mdx` - Home page
- `quickstart.mdx` - Getting started guide
- `api-reference/` - API documentation
  - `authentication.mdx` - Authentication documentation
  - `endpoint/` - Endpoint-specific documentation
- `openapi.json` - OpenAPI specification for API playground
- `logo/` - Logo assets

## Deployment

The documentation is automatically deployed when changes are pushed to the main branch.

## Version Support

Current version: 0.1.0

## API Playground

The API playground allows you to try out the API endpoints directly from the documentation. To use it:

1. Navigate to any API endpoint page
2. Enter your API key in the authentication field
3. Fill in the required parameters
4. Click "Try it" to send the request

## Customization

To customize the documentation:

1. Edit `mint.json` to change the configuration
2. Update the MDX files to change the content
3. Modify `openapi.json` to update the API specification
