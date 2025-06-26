#!/bin/bash

# Fix Java SDK compilation issues
echo "Fixing Java SDK compilation issues..."

# Add getter methods to ComputeHiveClient
echo "Adding missing getter methods to ComputeHiveClient..."
cat >> src/main/java/io/computehive/sdk/ComputeHiveClient.java << 'EOF'

    public String getApiUrl() {
        return apiUrl;
    }
    
    public OkHttpClient getHttpClient() {
        return httpClient;
    }
    
    public Gson getGson() {
        return gson;
    }
    
    public String getAccessToken() {
        return accessToken;
    }
}
EOF

# Remove the last closing brace and add getters
sed -i '' '$d' src/main/java/io/computehive/sdk/ComputeHiveClient.java

# Add simple logging
echo "Replacing Lombok logging with simple System.out..."
find src/main/java -name "*.java" -exec sed -i '' 's/log\.debug/System.out.println/g' {} \;
find src/main/java -name "*.java" -exec sed -i '' 's/log\.info/System.out.println/g' {} \;
find src/main/java -name "*.java" -exec sed -i '' 's/log\.warn/System.err.println/g' {} \;
find src/main/java -name "*.java" -exec sed -i '' 's/log\.error/System.err.println/g' {} \;

echo "Fix script completed. You may need to manually add constructors and getters to model classes."
echo "For production use, consider using a proper logging framework like SLF4J with Logback." 