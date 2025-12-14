import { GoogleGenAI } from "@google/genai";

// Initialize the API client
const ai = new GoogleGenAI({ apiKey: process.env.API_KEY || '' });

export const analyzeConfiguration = async (configContent: string, configType: string): Promise<string> => {
  if (!process.env.API_KEY) {
    throw new Error("API Key is missing. Please configure your environment variables to use AI analysis.");
  }

  try {
    const prompt = `
      You are a Senior DevOps Engineer. Analyze the following ${configType} configuration file.
      
      Please provide:
      1. A brief summary of what this service does based on the config.
      2. Identify any potential security risks (e.g., exposed ports, default passwords, root privileges).
      3. Suggest 1-2 optimizations or best practices.
      
      Keep the response concise and formatted in Markdown.
      
      Configuration:
      \`\`\`
      ${configContent}
      \`\`\`
    `;

    const response = await ai.models.generateContent({
      model: 'gemini-2.5-flash',
      contents: prompt,
    });

    if (!response.text) {
        throw new Error("The model returned an empty response.");
    }

    return response.text;
  } catch (error: any) {
    console.error("Error analyzing config:", error);
    // Extract error message if available
    const message = error instanceof Error ? error.message : "Unknown API error";
    throw new Error(`Analysis failed: ${message}`);
  }
};
