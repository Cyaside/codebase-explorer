import type { ConnectionProfile, SupportedProviderOption } from "@/lib/types";

export function validateProfile(
  profile: ConnectionProfile,
  apiKey: string,
  supportedProviders: SupportedProviderOption[],
) {
  const errors: string[] = [];
  const providerName = profile.provider.name.trim();
  if (!providerName) {
    errors.push("Provider mode is required for an AI-enabled connection.");
    return errors;
  }

  const descriptor = supportedProviders.find((item) => item.name === providerName);
  if (!descriptor) {
    errors.push(`Provider "${providerName}" is not supported by the local backend.`);
    return errors;
  }

  if (descriptor.requires_model && !profile.provider.model.trim()) {
    errors.push("Model is required for the selected provider.");
  }

  if (descriptor.requires_api_key && !apiKey.trim()) {
    errors.push("API key is required for the selected provider.");
  }

  if (descriptor.requires_base_url && !profile.provider.baseUrl.trim()) {
    errors.push("Base URL is required for the selected provider.");
  }

  return errors;
}
