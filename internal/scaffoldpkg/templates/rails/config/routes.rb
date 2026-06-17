Rails.application.routes.draw do
  get '/healthz', to: 'health#index'
end
