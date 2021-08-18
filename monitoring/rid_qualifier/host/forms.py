import json
from flask_wtf import FlaskForm
from wtforms import StringField, SubmitField, TextAreaField, BooleanField, IntegerField
from wtforms.validators import DataRequired, ValidationError
from wtforms.widgets import HiddenInput


class UserConfig(FlaskForm):
    auth_spec = StringField('Auth Spec', validators=[DataRequired()])
    user_config = TextAreaField('User Config', validators=[DataRequired()])
    sample_report = BooleanField('Sample Report')
    file_count = IntegerField(widget=HiddenInput())
    submit = SubmitField('Submit')

    def validate_user_config(form, field):
        user_config = json.loads(field.data)
        expected_keys = {'injection_targets', 'observers'}
        if not expected_keys.issubset(set(user_config)):
            message = f'missing fields in config object {expected_keys - set(user_config)}'
            raise ValidationError(message)
        if int(form.file_count.data) < len(user_config['injection_targets']):
            raise ValidationError(
                'Not enough flight states files provided for each injection_targets.')
