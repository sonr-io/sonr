const converter = require('swagger2openapi');
const glob = require('glob');
const fs = require('fs').promises;
const yaml = require('yaml');

/**
 * Script to find and convert Swagger 2.0 YAML files to OpenAPI 3.0 in-place.
 */
async function convertAll() {
  // Define options for the swagger2openapi converter.
  // Set 'patch' to true to perform some cleanup and fixing operations.
  const options = { patch: true, warnOnly: true };

  try {
    // 1. Find all files ending with .swagger.yaml in the target directory.
    const swaggerFiles = glob.sync('docs/static/openapi/*.swagger.yaml');

    if (swaggerFiles.length === 0) {
      console.log(
        "Conversion script finished: No *.swagger.yaml files were found in 'docs/static/openapi/'."
      );
      return;
    }

    console.log(`Found ${swaggerFiles.length} Swagger file(s) to convert.`);

    // 2. Create a list of conversion promises.
    const conversionPromises = swaggerFiles.map(async (filePath) => {
      try {
        console.log(`- Converting ${filePath}...`);

        // 3. Convert the file. The result object contains the OpenAPI definition.
        const { openapi } = await converter.convertFile(filePath, options);

        // 4. Serialize the resulting OpenAPI object back to a YAML string.
        const openapiYamlString = yaml.stringify(openapi);

        // 5. Overwrite the original file with the new OpenAPI 3.0 content.
        await fs.writeFile(filePath, openapiYamlString, 'utf8');

        console.log(`  √ Successfully converted and overwritten ${filePath}`);
      } catch (err) {
        console.error(`  × Failed to convert ${filePath}:`, err.message);
      }
    });

    // 6. Wait for all file conversions to complete.
    await Promise.all(conversionPromises);

    console.log('\nConversion process complete.');
  } catch (error) {
    console.error('\nAn unexpected error occurred:', error);
    process.exit(1); // Exit with an error code
  }
}

// Run the conversion process.
convertAll();
