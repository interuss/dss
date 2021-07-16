import json
from flask_wtf import FlaskForm
from wtforms import StringField, SubmitField, TextAreaField
from wtforms.validators import DataRequired, ValidationError

# {
#   "locale": "che",
#   "injection_targets": [
#     {
#       "name": "uss1",
#       "injection_base_url": "http://host.docker.internal:8070/sp/uss1"
#     },
#     {
#       "name": "uss2",
#       "injection_base_url": "http://host.docker.internal:8070/sp/uss2"
#     }
#   ],
#   "observers": [
#     {
#       "name": "uss2",
#       "observation_base_url": "http://host.docker.internal:8070/dp/uss2"
#     },
#     {
#       "name": "uss3",
#       "observation_base_url": "http://host.docker.internal:8070/dp/uss3"
#     }
#   ]
# }

class UserConfig(FlaskForm):
    auth_spec = StringField('Auth Spec', validators=[DataRequired()])
    user_config = TextAreaField('User Config', validators=[DataRequired()])
    submit = SubmitField('Submit')

    def validate_user_config(form, field):
        try:
            user_config = json.loads(field.data)
        except ValueError as e:
            raise ValueError(e)
        expected_keys = {'locale', 'injection_targets', 'observers'}
        if not expected_keys.issubset(set(user_config)):
            message = f'missing fields in config object {expected_keys - set(user_config)}'
            raise ValidationError(message)
